# SSH Key Setup Guide for GitHub Actions Deployment

This guide walks through setting up SSH keys for automated deployment via GitHub Actions CI/CD pipeline.

## Quick Reference

```bash
# Generate SSH key pair
ssh-keygen -t rsa -b 4096 -f deploy_key -N ""

# Copy to authorized_keys on deployment server
cat deploy_key.pub | ssh user@production.example.com "cat >> ~/.ssh/authorized_keys"

# Add to GitHub Secrets
# Go to: Settings → Secrets and variables → Actions
# Create secret: DEPLOY_KEY with contents of deploy_key file
```

## Step-by-Step Setup

### Step 1: Generate SSH Key Pair (Local Machine)

```bash
# Generate 4096-bit RSA key pair (recommended for security)
ssh-keygen -t rsa -b 4096 -f ./deploy_key -N ""

# This creates:
# - deploy_key (private key - keep secret!)
# - deploy_key.pub (public key - share with servers)

# Verify keys were created
ls -la deploy_key*
```

### Step 2: Configure Deployment Server

**Option A: Manual Setup (One-time)**

```bash
# SSH to production server
ssh ubuntu@production.example.com

# Create .ssh directory if not exists
mkdir -p ~/.ssh
chmod 700 ~/.ssh

# Add public key to authorized_keys
echo "PUBLIC_KEY_CONTENT" >> ~/.ssh/authorized_keys
chmod 600 ~/.ssh/authorized_keys

# Test connection (from local machine)
ssh -i deploy_key ubuntu@production.example.com "echo 'SSH works!'"
```

**Option B: Automated Setup (Using script)**

```bash
# Save as setup_deploy_key.sh
#!/bin/bash
set -e

DEPLOY_USER=${1:-ubuntu}
DEPLOY_HOST=${2:-production.example.com}
PUBLIC_KEY_FILE=${3:-deploy_key.pub}

echo "Setting up SSH key for $DEPLOY_USER@$DEPLOY_HOST..."

# Copy public key to server
ssh-copy-id -i "$PUBLIC_KEY_FILE" "$DEPLOY_USER@$DEPLOY_HOST"

# Verify
ssh -i deploy_key "$DEPLOY_USER@$DEPLOY_HOST" "echo 'SSH configured successfully!'"

echo "✅ SSH key setup complete"

# Usage:
# ./setup_deploy_key.sh ubuntu production.example.com deploy_key.pub
```

### Step 3: Add SSH Key to GitHub Secrets

**Via GitHub Web UI (Easiest):**

1. Go to your repository on GitHub
2. Click "Settings" tab
3. In left sidebar, click "Secrets and variables" → "Actions"
4. Click "New repository secret" button
5. **Name**: `DEPLOY_KEY`
6. **Value**: Paste entire contents of `deploy_key` file (private key)
   ```bash
   cat deploy_key
   ```
7. Click "Add secret"

**Via GitHub CLI (Alternative):**

```bash
# Install GitHub CLI first
# https://cli.github.com

# Login to GitHub
gh auth login

# Add secret
gh secret set DEPLOY_KEY < deploy_key

# List secrets
gh secret list

# Delete secret (if needed)
gh secret delete DEPLOY_KEY
```

### Step 4: Add Other Required Secrets

```bash
# Via GitHub Web UI, add these secrets:

DEPLOY_USER          = ubuntu
STAGING_HOST         = staging.example.com
PROD_HOST            = production.example.com
DEPLOY_PATH          = /home/ubuntu/sovereign-map
SLACK_WEBHOOK        = https://hooks.slack.com/services/... (optional)
```

### Step 5: Test SSH Connection

```bash
# Test locally (to verify key works)
ssh -i deploy_key ubuntu@production.example.com "docker --version"

# Should return Docker version without prompting for password
```

### Step 6: Configure GitHub Actions Workflow

The workflow file `.github/workflows/deploy.yml` is already configured.

Check these fields match your setup:

```yaml
- name: Deploy to production via SSH
  uses: appleboy/ssh-action@master
  with:
    host: ${{ secrets.PROD_HOST }}              # From GitHub Secret
    username: ${{ secrets.DEPLOY_USER }}        # From GitHub Secret
    key: ${{ secrets.DEPLOY_KEY }}              # From GitHub Secret (private key)
    script: |
      # SSH commands execute here
```

### Step 7: Verify GitHub Actions Access

**In GitHub Repository:**

1. Go to "Actions" tab
2. Select "Build & Deploy" workflow
3. Click "Run workflow" → "Run workflow"
4. Watch the deployment progress
5. Check "Deploy to production" step for SSH connection

**Expected Output:**

```
SSH key successfully connected ✅
Current directory: /home/ubuntu/sovereign-map
Docker containers deployed: 100 nodes
Deployment complete: http://production.example.com:8000
```

## Troubleshooting SSH Connection

### Problem: "Permission denied (publickey)"

**Solution 1: Verify public key on server**

```bash
# SSH to server and check
ssh ubuntu@production.example.com

# List authorized keys
cat ~/.ssh/authorized_keys

# Should contain the exact public key content
cat ~/.ssh/authorized_keys | grep "$(cat deploy_key.pub | awk '{print $NF}')"
```

**Solution 2: Verify private key permissions**

```bash
# Local machine - check key permissions
ls -la deploy_key
# Should show: -rw------- (600 permissions)

# Fix if needed
chmod 600 deploy_key
```

**Solution 3: Test direct connection**

```bash
# Add verbose output to debug
ssh -vvv -i deploy_key ubuntu@production.example.com

# Look for error messages
# Common issues:
# - "No public key found" → Key not in authorized_keys
# - "Too many authentication failures" → Wrong key or permissions
# - "Connection refused" → SSH daemon not running
```

### Problem: "Host key verification failed"

**Solution: Add host key to known_hosts**

```bash
# Add server to known_hosts
ssh-keyscan -H production.example.com >> ~/.ssh/known_hosts

# Or disable host key checking (not recommended)
ssh -o StrictHostKeyChecking=no -i deploy_key ubuntu@production.example.com
```

### Problem: GitHub Actions workflow fails with SSH error

**Solution: Check GitHub Actions logs**

1. Go to "Actions" tab in GitHub
2. Click on the failed workflow run
3. Click on "Deploy to production" job
4. Look for SSH error messages
5. Common issues:
   - DEPLOY_KEY secret not set or corrupted
   - PROD_HOST not reachable (firewall issue)
   - SSH user doesn't have docker permissions

**Fix Docker Permissions:**

```bash
# SSH to server
ssh ubuntu@production.example.com

# Add ubuntu user to docker group
sudo usermod -aG docker ubuntu

# Apply group changes
newgrp docker

# Test docker access
docker ps

# Logout and login again
exit
ssh ubuntu@production.example.com "docker ps"
```

## Security Best Practices

### 1. Generate New Key for Each Environment

```bash
# Production key
ssh-keygen -t rsa -b 4096 -f deploy_key_prod -N ""

# Staging key
ssh-keygen -t rsa -b 4096 -f deploy_key_staging -N ""

# Use appropriate key for each environment
```

### 2. Rotate Keys Regularly

```bash
# Generate new key
ssh-keygen -t rsa -b 4096 -f deploy_key_new -N ""

# Add new public key to server
ssh-copy-id -i deploy_key_new.pub ubuntu@production.example.com

# Remove old key from server
ssh ubuntu@production.example.com "sed -i '/old_key_fingerprint/d' ~/.ssh/authorized_keys"

# Update GitHub Secret with new private key
gh secret set DEPLOY_KEY < deploy_key_new
```

### 3. Restrict SSH Permissions (Server-side)

```bash
# SSH to server and edit authorized_keys
ssh ubuntu@production.example.com

# Edit and add restrictions to the key
# Before: ssh-rsa AAAA...
# After: command="docker-compose pull && docker-compose up -d", restrict ssh-rsa AAAA...

# This limits what the key can do
```

### 4. Monitor SSH Access

```bash
# Check SSH logs on server
sudo tail -f /var/log/auth.log | grep sshd

# Look for:
# - Successful connections: "Accepted publickey"
# - Failed attempts: "Failed publickey"
# - Invalid users: "Invalid user"
```

### 5. Never Commit Private Keys

```bash
# Add to .gitignore
echo "deploy_key" >> .gitignore
echo "deploy_key_*" >> .gitignore
echo "*_key" >> .gitignore

# Verify not committed
git log --all --full-history -- deploy_key

# If accidentally committed, regenerate key immediately!
```

## Advanced Configuration

### Using SSH Agent (Local Development)

```bash
# Start SSH agent
eval "$(ssh-agent -s)"

# Add key
ssh-add deploy_key

# Test (should not prompt for password)
ssh ubuntu@production.example.com "docker ps"

# Check loaded keys
ssh-add -l

# Remove key
ssh-add -d deploy_key
```

### Using SSH Config (Local Development)

Create `~/.ssh/config`:

```
Host production
    HostName production.example.com
    User ubuntu
    IdentityFile ~/.ssh/deploy_key
    IdentitiesOnly yes

Host staging
    HostName staging.example.com
    User ubuntu
    IdentityFile ~/.ssh/deploy_key_staging
    IdentitiesOnly yes
```

Then use:

```bash
# Instead of:
ssh -i deploy_key ubuntu@production.example.com

# Just:
ssh production

# And in GitHub Actions, use hostname
ssh production "docker ps"
```

### Using GitHub Deploy Keys (Repository-level)

Alternative to personal secrets (not recommended for this use case):

```bash
# 1. Go to repository Settings
# 2. Click "Deploy keys"
# 3. Add your deploy_key.pub
# 4. Check "Allow write access"
# 5. GitHub will automatically rotate keys if needed
```

## Firewall & Network Considerations

### Open SSH Port

```bash
# Ubuntu/Debian firewall
sudo ufw allow 22/tcp
sudo ufw allow 22

# AWS Security Group
# Add inbound rule:
# Type: SSH (22)
# Protocol: TCP
# Port: 22
# Source: GitHub Actions runner IPs (0.0.0.0/0 or restricted)

# GCP Firewall
gcloud compute firewall-rules create allow-ssh \
  --allow tcp:22 \
  --source-ranges 0.0.0.0/0
```

### Test Port Connectivity

```bash
# From GitHub Actions (uses their runner IPs)
# These should be reachable

# From your local machine:
nc -zv production.example.com 22
# or
telnet production.example.com 22
```

## Post-Deployment Verification

### 1. Check Deployment Logs

```bash
# On production server
cd /home/ubuntu/sovereign-map

# View deployment history
git log --oneline -5

# Check current containers
docker-compose ps

# Tail logs
docker-compose logs -f backend
```

### 2. Monitor GitHub Actions

```bash
# Via GitHub web UI
# Settings → Actions → General
# View workflow runs and logs

# Via GitHub CLI
gh run list --workflow=deploy.yml

# View specific run
gh run view <run-id> --log
```

### 3. Validate Application

```bash
# From production server
curl http://localhost:8000/health
curl http://localhost:8000/convergence

# Check metrics
curl http://localhost:9090/api/v1/targets

# View Grafana
# http://production.example.com:3000
```

## Cleanup

### Remove SSH Keys

```bash
# After deployment is stable, clean up local keys

# List keys
ls -la deploy_key*

# Remove (optional, can keep as backup)
rm deploy_key deploy_key.pub

# Verify not committed
git status
git log --all --full-history -- deploy_key
```

### Rotate Keys in Production

```bash
# 1. Generate new keys
ssh-keygen -t rsa -b 4096 -f deploy_key_new -N ""

# 2. Add new key to server
ssh-copy-id -i deploy_key_new.pub ubuntu@production.example.com

# 3. Update GitHub Secret
gh secret set DEPLOY_KEY < deploy_key_new

# 4. Test new key in GitHub Actions
# Push a test commit and watch workflow

# 5. Remove old key from authorized_keys (optional)
ssh ubuntu@production.example.com "sed -i '/old_fingerprint/d' ~/.ssh/authorized_keys"

# 6. Remove old local keys
rm deploy_key*
```

## Summary Checklist

- [ ] Generated SSH key pair (4096-bit RSA)
- [ ] Added public key to ~/.ssh/authorized_keys on production server
- [ ] Tested SSH connection locally
- [ ] Added DEPLOY_KEY secret to GitHub Actions
- [ ] Added PROD_HOST, DEPLOY_USER, DEPLOY_PATH secrets
- [ ] GitHub Actions workflow file exists (.github/workflows/deploy.yml)
- [ ] Verified workflow can access secrets
- [ ] Tested deployment with manual workflow trigger
- [ ] Verified application running on production
- [ ] Set up monitoring for SSH access
- [ ] Documented key location for team
- [ ] Configured key rotation schedule

---

**Last Updated**: February 2024  
**SSH Protocol**: RSA 4096-bit  
**Encryption**: AES-128-GCM (standard GitHub Secret encryption)  
**Status**: Production Ready ✅
