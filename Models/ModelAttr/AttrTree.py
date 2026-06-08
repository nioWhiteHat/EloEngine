import argparse
import os
import re
import onnxmltools
import pandas as pd
import xgboost as xgb
from skl2onnx.common.data_types import FloatTensorType

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--input_csv', type=str, required=True)
    parser.add_argument('--output_json', type=str, required=True)
    parser.add_argument('--target_col', type=str, default='AskingPricePerSqm')
    parser.add_argument('--columns_dropped', type=str, default='UID,GeoLat,GeoLon,Distance,AgeSq,InteractionAgeDistance,LogInteraction,AskingPrice,AdCreationAge')
    args = parser.parse_args()

    df = pd.read_csv(args.input_csv)

    drop_cols = [c.strip() for c in args.columns_dropped.split(',')]
    drop_cols_present = [c for c in drop_cols if c in df.columns]

    output_df = pd.DataFrame()
    for col in ['UID', 'GeoLat', 'GeoLon', args.target_col]:
        if col in df.columns:
            output_df[col] = df[col]

    X = df.drop(columns=drop_cols_present + [args.target_col])
    y = df[args.target_col]

    categorical_cols = X.select_dtypes(include=['object', 'string']).columns
    with open("Models/Saved/CategoryMappings.txt", "w", encoding="utf-8") as f:
        for col in categorical_cols:
            X[col] = X[col].astype('category')
            for i, cat in enumerate(X[col].cat.categories):
                f.write(f"{col},{cat},{i}\n")

    model = xgb.XGBRegressor(
        enable_categorical=True,
        tree_method='hist',
        max_depth=8,
        min_child_weight=30,
        learning_rate=0.3,
        n_estimators=15,
        subsample=0.8,
        colsample_bytree=0.8
    )
    X.columns = [f"f{i}" for i in range(X.shape[1])]
    model.fit(X, y)

    preds = model.predict(X)

    output_df['predicted_sqm_price'] = preds
    output_df['residual_percentage'] = (output_df[args.target_col] - preds) / preds

    clean_path = re.sub(r'[<>:"|?*\x00-\x1f\']', '', args.output_json).strip()
    clean_path = os.path.abspath(clean_path)
    os.makedirs(os.path.dirname(clean_path), exist_ok=True)

    output_df.to_json(clean_path, orient='records', indent=2)
    print("SUCCESS")
    print("Converting model to ONNX format...")
    # We define a single float tensor for all 30 inputs. 
    # Categoricals (encoded as ints) will be passed as floats, which ONNX handles natively for XGBoost.
    num_features = X.shape[1]
    initial_types = [('input', FloatTensorType([None, num_features]))]
    
    onnx_model = onnxmltools.convert_xgboost(model, initial_types=initial_types)
    
    # Save the clean ONNX file
    onnx_file = "Models/Saved/ModelAttr_Trees.onnx"
    with open(onnx_file, "wb") as f:
        f.write(onnx_model.SerializeToString())
    print(f"SUCCESS: Saved {onnx_file}")
if __name__ == "__main__":
    main()