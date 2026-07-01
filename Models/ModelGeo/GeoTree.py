import argparse
import pandas as pd
import xgboost as xgb

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--input_csv', type=str, required=True)
    parser.add_argument('--input_json', type=str, required=True)
    parser.add_argument('--output_json', type=str, required=True)
    parser.add_argument('--target_col', type=str, default='AskingPricePerSqm')
    parser.add_argument('--columns_dropped', type=str, default='UID,GeoLat,GeoLon,Distance,AgeSq,InteractionAgeDistance,LogInteraction,AskingPrice,AdCreationAge,ZoneValue,MarketabilityFactor')
    args = parser.parse_args()

    features_df = pd.read_csv(args.input_csv)
    geo_df = pd.read_json(args.input_json)

    geo_subset = geo_df[['UID', 'geo_premium']]
    df = pd.merge(features_df, geo_subset, on='UID', how='inner')

    drop_cols = [c.strip() for c in args.columns_dropped.split(',')]
    drop_cols_present = [c for c in drop_cols if c in df.columns]

    output_df = pd.DataFrame()
    for col in ['UID', 'GeoLat', 'GeoLon', args.target_col, 'geo_premium']:
        if col in df.columns:
            output_df[col] = df[col]

    X = df.drop(columns=drop_cols_present + [args.target_col])
    y = df[args.target_col]

    categorical_cols = X.select_dtypes(include=['object', 'string']).columns
    for col in categorical_cols:
        X[col] = X[col].astype('category')

    model = xgb.XGBRegressor(
        enable_categorical=True,
        tree_method='hist',
        max_depth=6,
        min_child_weight=15,
        learning_rate=0.05,
        n_estimators=500,
        subsample=0.8,
        colsample_bytree=0.8
    )

    model.fit(X, y)
    preds = model.predict(X)

    output_df['final_predicted_sqm_price'] = preds
    output_df['final_residual_percentage'] = (output_df[args.target_col] - preds) / preds

    output_df.to_json(args.output_json, orient='records', indent=2)
    print("SUCCESS")
    model.save_model("Models/Saved/ModelGeo_Trees.json")
if __name__ == "__main__":
    main()