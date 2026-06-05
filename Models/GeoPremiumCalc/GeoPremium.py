import argparse
import pandas as pd
import numpy as np
from sklearn.neighbors import NearestNeighbors

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--input_json', type=str, required=True)
    parser.add_argument('--output_json', type=str, required=True)
    parser.add_argument('--max_neighbors', type=int, default=25) # Max houses to average
    parser.add_argument('--radius_meters', type=float, default=200.0) # Absolute distance limit
    args = parser.parse_args()

    df = pd.read_json(args.input_json)
    residuals = df['residual_percentage'].values

    # 1. Haversine requires coordinates in radians
    coords_rad = np.radians(df[['GeoLat', 'GeoLon']].values)
    radius_rad = args.radius_meters / 6371000.0

    # We search for max_neighbors + 1, because the property will always find ITSELF 
    # at distance 0.0, and we want 15 *other* houses.
    k_search = args.max_neighbors + 1 

    # 2. Find the absolute closest neighbors, ignoring distance for a moment
    knn = NearestNeighbors(n_neighbors=k_search, metric='haversine')
    knn.fit(coords_rad)
    distances, indices = knn.kneighbors(coords_rad)

    # 3. Apply the Hybrid Filter
    geo_premiums = []
    for i in range(len(df)):
        prop_dists = distances[i]
        prop_idxs = indices[i]
        
        # Keep only neighbors that are within the 500m radius AND are not the target property itself (dist > 0)
        valid_mask = (prop_dists <= radius_rad) & (prop_dists > 0.0)
        valid_neighbors = prop_idxs[valid_mask]
        
        # Calculate the average of whoever survived the filter
        if len(valid_neighbors) > 0:
            avg_premium = np.mean(residuals[valid_neighbors])
            geo_premiums.append(avg_premium)
        else:
            geo_premiums.append(0.0) # Fallback if zero houses are within 500m

    df['geo_premium'] = geo_premiums
    df.to_json(args.output_json, orient='records', indent=2)

if __name__ == "__main__":
    main()