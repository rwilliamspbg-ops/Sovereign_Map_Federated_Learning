🚀 Quick Start

```bash
Copy
# Download all files, then:
chmod +x *.sh
./run-test.sh all
```
---
Or step by step:
```bash
Copy
./run-test.sh setup    # Phase 1: AWS setup
./run-test.sh deploy   # Phase 2: Deploy 200 EC2 instances
./run-test.sh code     # Phase 3: Deploy FL code
./run-test.sh test     # Phase 4: Run 3-hour training
./run-test.sh results  # Phase 5: Capture results
./run-test.sh cleanup  # Phase 6: Destroy infrastructure
```
---
All files are ready to use for a legitimate 200-node federated learning test on AWS with Byzantine fault tolerance and differential privacy.
