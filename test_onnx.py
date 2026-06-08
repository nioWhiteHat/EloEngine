import sys
import traceback

log_file = "debug_log.txt"

def log(msg):
    with open(log_file, "a", encoding="utf-8") as f:
        f.write(msg + "\n")

# Clear the log file before starting
with open(log_file, "w", encoding="utf-8") as f:
    f.write("--- STARTING ONNX TEST ---\n")

try:
    log("STEP 1: Imports starting...")
    import onnxruntime as rt
    import numpy as np
    log("STEP 1: Imports worked.")

    log("STEP 2: Loading model...")
    sess = rt.InferenceSession("Models/Saved/ModelAttr_Trees.onnx")
    log("STEP 2: Model loaded.")
    
    input_name = sess.get_inputs()[0].name
    
    go_array = [
        126.0, 6.0, 1.0, 0.0, 118.0, 118.0, 0.0, 0.0, 0.0, 0.0, 0.0,
        0.0, 1.0, 1.0, 0.0, 1.0, 0.0, 0.0, 0.0, 2.0, 1.0, 0.0, 0.0,
        0.0, 1.0, 3.0, 0.0, 0.0, 1.0, 1.0, 1850.0, 1.0, 0.0
    ]
    
    data = np.array([go_array], dtype=np.float32)
    log("STEP 3: Array built, running prediction...")
    
    pred = sess.run(None, {input_name: data})
    
    log("STEP 4: SUCCESS! Final Price:")
    log(str(pred[0][0]))

except Exception as e:
    log(f"FATAL ERROR: {str(e)}")
    log(traceback.format_exc())