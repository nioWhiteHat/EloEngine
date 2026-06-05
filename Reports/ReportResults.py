import argparse
import pandas as pd

def generate_html(input_csv, output_html):
    df = pd.read_csv(input_csv)

    df['MicroNeighborhood_Key'] = df['GeoPremium'].round(3)

    df = df.sort_values(
        by=['MicroNeighborhood_Key', 'FinalResidual'], 
        ascending=[False, True]
    )

    html_content = """
    <!DOCTYPE html>
    <html lang="en">
    <head>
        <meta charset="UTF-8">
        <title>Arbitrage Report</title>
        <style>
            body { font-family: 'Segoe UI', system-ui, sans-serif; background: #0f172a; color: #e2e8f0; padding: 20px; margin: 0; }
            h1 { color: #f8fafc; margin-bottom: 30px; font-weight: 600; padding-left: 10px; }
            .group-card { background: #1e293b; border-radius: 12px; padding: 20px; margin-bottom: 30px; border-left: 5px solid #3b82f6; box-shadow: 0 10px 15px -3px rgba(0,0,0,0.3); }
            .group-title { margin-top: 0; color: #93c5fd; font-size: 1.25rem; font-weight: 600; margin-bottom: 20px; }
            
            .grid-header { display: grid; grid-template-columns: 1.5fr 1.5fr 1.5fr 1fr 1fr 0.5fr; gap: 15px; padding: 10px 20px; font-weight: 600; color: #94a3b8; font-size: 0.85rem; text-transform: uppercase; letter-spacing: 0.05em; border-bottom: 1px solid #334155; }
            
            .prop-item { background: #0f172a; border-radius: 8px; margin-bottom: 8px; border: 1px solid #334155; overflow: hidden; }
            .prop-summary { display: grid; grid-template-columns: 1.5fr 1.5fr 1.5fr 1fr 1fr 0.5fr; gap: 15px; padding: 16px 20px; cursor: pointer; align-items: center; list-style: none; transition: background 0.2s ease; }
            .prop-summary::-webkit-details-marker { display: none; }
            .prop-summary:hover { background: #1e293b; }
            .prop-summary:focus { outline: none; }
            
            .val-deal { color: #4ade80; font-weight: 700; font-size: 1.1rem; }
            .val-over { color: #f87171; font-weight: 700; font-size: 1.1rem; }
            .metric { font-size: 1.05rem; }
            .expand-icon { color: #64748b; text-align: right; font-size: 1.2rem; transition: transform 0.2s; }
            
            details[open] .expand-icon { transform: rotate(180deg); }
            details[open] .prop-summary { border-bottom: 1px solid #334155; background: #1e293b; }
            
            .extended-data { display: grid; grid-template-columns: repeat(auto-fill, minmax(220px, 1fr)); gap: 16px; padding: 24px; background: #0f172a; }
            .attr-box { background: #1e293b; padding: 12px 16px; border-radius: 6px; border: 1px solid #334155; }
            .attr-label { color: #64748b; display: block; font-size: 0.7rem; text-transform: uppercase; letter-spacing: 0.5px; margin-bottom: 4px; }
            .attr-value { color: #f8fafc; font-weight: 600; font-size: 0.95rem; word-break: break-all; }
        </style>
    </head>
    <body>
        <h1>Geospatial Arbitrage Report</h1>
    """

    groups = df.groupby('MicroNeighborhood_Key', sort=False)

    for premium, group in groups:
        html_content += f"""
        <div class="group-card">
            <h2 class="group-title">Micro-Neighborhood GeoPremium: {premium * 100:+.2f}%</h2>
            <div class="grid-header">
                <div>Final Residual</div>
                <div>Predicted €/Sqm</div>
                <div>Actual €/Sqm</div>
                <div>Age</div>
                <div>Sqm</div>
                <div></div>
            </div>
        """

        for _, row in group.iterrows():
            res_val = row['FinalResidual']
            res_class = "val-deal" if res_val < 0 else "val-over"
            res_text = f"{res_val * 100:+.1f}%"

            extended_html = "<div class='extended-data'>"
            skip_cols = {'FinalResidual', 'GeoPremium', 'PredictedPricePerSqm', 'MicroNeighborhood_Key', 'Age', 'SquareMeters', 'AskingPricePerSqm'}
            
            for col in df.columns:
                if col not in skip_cols:
                    extended_html += f"""
                    <div class="attr-box">
                        <span class="attr-label">{col}</span>
                        <span class="attr-value">{row[col]}</span>
                    </div>
                    """
            extended_html += "</div>"

            html_content += f"""
            <details class="prop-item">
                <summary class="prop-summary">
                    <div class="{res_class}">{res_text}</div>
                    <div class="metric">€{row['PredictedPricePerSqm']:,.0f}</div>
                    <div class="metric">€{row['AskingPricePerSqm']:,.0f}</div>
                    <div class="metric">{row['Age']}</div>
                    <div class="metric">{row['SquareMeters']}</div>
                    <div class="expand-icon">▼</div>
                </summary>
                {extended_html}
            </details>
            """

        html_content += """
        </div>
        """

    html_content += "</body></html>"

    with open(output_html, 'w', encoding='utf-8') as f:
        f.write(html_content)

    print(f"SUCCESS: Report saved to {output_html}")

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument('--input_csv', type=str, required=True)
    parser.add_argument('--output_html', type=str, required=True)
    args = parser.parse_args()
    generate_html(args.input_csv, args.output_html)