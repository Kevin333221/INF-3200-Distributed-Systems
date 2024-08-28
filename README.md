# How to run the code

1. Ensure run.sh and clean.sh have execute permissions. If not, run the following commands:
    ```bash
    chmod +x run.sh
    chmod +x clean.sh
    ```
    
2. Run:
    ```bash
    ./run.sh
    ```
    
3. Use the output array to run the testscript. The array needs to be enclosed in single quotes:
    ```bash
    python3 testscript.py '<output_array>'
    ```