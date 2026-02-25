🚀 Quick Start
| File                                                                                                                           | Size         | Description                     | Download                                                                                           |
| ------------------------------------------------------------------------------------------------------------------------------ | ------------ | ------------------------------- | -------------------------------------------------------------------------------------------------- |
| **[run-test.sh](sandbox:///mnt/kimi/output/sovereign-fl-200-node-test/run-test.sh)**                                           | 3,053 bytes  | **Master orchestration script** | [Download](sandbox:///mnt/kimi/output/sovereign-fl-200-node-test/run-test.sh)                      |
| **[phase-1-aws-setup.sh](sandbox:///mnt/kimi/output/sovereign-fl-200-node-test/phase-1-aws-setup.sh)**                         | 8,198 bytes  | AWS account setup               | [Download](sandbox:///mnt/kimi/output/sovereign-fl-200-node-test/phase-1-aws-setup.sh)             |
| **[phase-2-deploy-infrastructure.sh](sandbox:///mnt/kimi/output/sovereign-fl-200-node-test/phase-2-deploy-infrastructure.sh)** | 15,808 bytes | Terraform deployment            | [Download](sandbox:///mnt/kimi/output/sovereign-fl-200-node-test/phase-2-deploy-infrastructure.sh) |
| **[phase-3-deploy-code.sh](sandbox:///mnt/kimi/output/sovereign-fl-200-node-test/phase-3-deploy-code.sh)**                     | 14,713 bytes | Code deployment                 | [Download](sandbox:///mnt/kimi/output/sovereign-fl-200-node-test/phase-3-deploy-code.sh)           |
| **[phase-4-execute-test.sh](sandbox:///mnt/kimi/output/sovereign-fl-200-node-test/phase-4-execute-test.sh)**                   | 10,156 bytes | Test execution                  | [Download](sandbox:///mnt/kimi/output/sovereign-fl-200-node-test/phase-4-execute-test.sh)          |
| **[phase-5-capture-results.sh](sandbox:///mnt/kimi/output/sovereign-fl-200-node-test/phase-5-capture-results.sh)**             | 18,062 bytes | Results capture                 | [Download](sandbox:///mnt/kimi/output/sovereign-fl-200-node-test/phase-5-capture-results.sh)       |
| **[phase-6-cleanup.sh](sandbox:///mnt/kimi/output/sovereign-fl-200-node-test/phase-6-cleanup.sh)**                             | 1,951 bytes  | Cleanup infrastructure          | [Download](sandbox:///mnt/kimi/output/sovereign-fl-200-node-test/phase-6-cleanup.sh)               |
| **[src/aggregator.py](sandbox:///mnt/kimi/output/sovereign-fl-200-node-test/src/aggregator.py)**                               | 4,396 bytes  | FL server with Multi-Krum       | [Download](sandbox:///mnt/kimi/output/sovereign-fl-200-node-test/src/aggregator.py)                |
| **[src/client.py](sandbox:///mnt/kimi/output/sovereign-fl-200-node-test/src/client.py)**                                       | 3,984 bytes  | FL client with DP-SGD           | [Download](sandbox:///mnt/kimi/output/sovereign-fl-200-node-test/src/client.py)                    |
| **[terraform/main.tf](sandbox:///mnt/kimi/output/sovereign-fl-200-node-test/terraform/main.tf)**                               | 3,404 bytes  | AWS infrastructure as code      | [Download](sandbox:///mnt/kimi/output/sovereign-fl-200-node-test/terraform/main.tf)                |
| **[QUICKSTART.md](sandbox:///mnt/kimi/output/sovereign-fl-200-node-test/QUICKSTART.md)**                                       | 1,151 bytes  | Quick start guide               | [Download](sandbox:///mnt/kimi/output/sovereign-fl-200-node-test/QUICKSTART.md)                    |
| **[README.md](sandbox:///mnt/kimi/output/sovereign-fl-200-node-test/README.md)**                                               | 1,002 bytes  | Documentation                   | [Download](sandbox:///mnt/kimi/output/sovereign-fl-200-node-test/README.md)                        |
--
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
