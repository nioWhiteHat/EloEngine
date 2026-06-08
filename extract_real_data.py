import pandas as pd
import traceback

log_file = "true_data_dump.txt"

def write_log(msg):
    with open(log_file, "a", encoding="utf-8") as f:
        f.write(msg + "\n")

with open(log_file, "w", encoding="utf-8") as f:
    f.write("")

try:
    df = pd.read_csv("Modeled_Sold.csv")
    
    target_uid = "property_01KP6BYMQS9Z0X49DPBG92KB69"
    target_row = df[df['UID'] == target_uid].copy()

    if target_row.empty:
        write_log("ERROR: Could not find UID in CSV.")
    else:
        columns_dropped = 'UID,GeoLat,GeoLon,Distance,AgeSq,InteractionAgeDistance,LogInteraction,AskingPrice,AdCreationAge'
        drop_cols = [c.strip() for c in columns_dropped.split(',')]
        drop_cols_present = [c for c in drop_cols if c in df.columns]
        target_col = 'AskingPricePerSqm'

        X_test = target_row.drop(columns=drop_cols_present + [target_col])

        categorical_cols = X_test.select_dtypes(include=['object', 'string']).columns
        
        for col in categorical_cols:
            X_test[col] = X_test[col].astype('category').cat.codes

        write_log("EXACT GO ARRAY:")
        write_log(str(X_test.iloc[0].values.tolist()))
        
        write_log("\nEXACT COLUMN ORDER:")
        for i, col in enumerate(X_test.columns):
            write_log(f"{i}: {col}")

except Exception as e:
    write_log(f"FAILED: {str(e)}")