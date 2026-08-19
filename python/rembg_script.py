import sys
import os
from rembg import remove, new_session
from PIL import Image

# Initialize the lightweight model session for 1GB VDS
session = new_session("u2netp")

def process_file(input_path, output_path):
    try:
        input_image = Image.open(input_path)
        output_image = remove(input_image, session=session)
        output_image.save(output_path)
        print(f"SUCCESS:{output_path}")
    except Exception as e:
        print(f"ERROR:{e}", file=sys.stderr)
        sys.exit(1)

if __name__ == "__main__":
    if len(sys.argv) != 3:
        print("Usage: python rembg_script.py <input> <output>", file=sys.stderr)
        sys.exit(1)
    
    input_p = sys.argv[1]
    output_p = sys.argv[2]
    
    # If it's a directory, process all files (useful for video frames)
    if os.path.isdir(input_p):
        os.makedirs(output_p, exist_ok=True)
        for filename in os.listdir(input_p):
            if filename.endswith(".png"):
                in_f = os.path.join(input_p, filename)
                out_f = os.path.join(output_p, filename)
                process_file(in_f, out_f)
    else:
        process_file(input_p, output_p)
