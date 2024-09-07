## **Neutaro Validator Security Guide**

### **Disclaimer: Proceed with Caution**

Before implementing any of the security configurations or commands mentioned in this guide, it is crucial that you fully understand the potential impact of these changes on your node and overall system. Improper configurations can lead to loss of access, data, or security vulnerabilities. 

**Important Points to Consider:**
- **Know What You’re Doing:** Always ensure you understand the commands you are executing and the changes being made to your system. If you are unsure about a step, take the time to research or consult with an expert.
- **Test in a Safe Environment:** If possible, test any changes in a controlled environment before applying them to your live validator or server.
- **We Do Not Take Responsibility:** The information provided in this guide is for educational purposes only. By following this guide, you acknowledge that any changes you make are at your own risk. Timpi or its community members are not liable for any damages, loss of funds, or disruptions that may occur as a result of implementing these instructions.

Always prioritize the security and stability of your node and validate actions carefully.

### **1. Secure Key Management**

#### **a. Use `keyring-backend os` for Secure Key Storage**
**What It Does:** Stores keys securely in the operating system’s keychain rather than plaintext files.

**How to Set It Up:**

1. **When creating or managing your validator, use the `--keyring-backend os` flag**:

    ```shell
    neutaro keys add YourValidatorKey --keyring-backend os
    ```

2. **Whenever you manage keys (e.g., viewing, deleting), always use `--keyring-backend os`:**

    ```shell
    neutaro keys list --keyring-backend os
    ```

#### **b. Never Store Seed Phrases Locally**
**Best Practice:** Keep seed phrases offline to avoid unauthorized access.

**How to Store Seeds Securely:**

- Write down the seed phrase on paper and store it in a safe place like a safe or safety deposit box.
- Consider using a hardware wallet to securely store seed phrases.
- Avoid digital storage (e.g., text files, cloud storage) to minimize risk.


### **2. Implement Strong Server Security Practices**

#### **a. Use Firewalls and Access Controls**

**What It Does:** Restricts access to only necessary services, reducing the attack surface.

**How to Set It Up:**

1. **Install UFW (Uncomplicated Firewall):**

    ```shell
    sudo apt-get install ufw
    ```

2. **Set up UFW to allow only necessary traffic:**

    ```shell
    sudo ufw default deny incoming  # Deny all incoming connections by default
    sudo ufw default allow outgoing  # Allow all outgoing connections by default
    sudo ufw allow ssh  # Allow SSH access for remote management
    sudo ufw allow 26656/tcp  # Allow your validator port (replace with your port)
    sudo ufw enable  # Enable the firewall
    ```

3. **Verify UFW status:**

    ```shell
    sudo ufw status
    ```

4. **Disable Root Login via SSH:**

   - **Edit SSH configuration file**:

    ```shell
    sudo vim /etc/ssh/sshd_config
    ```

   - **Set `PermitRootLogin no`** to disable root login.
   - **Restart SSH to apply changes:**

    ```shell
    sudo systemctl restart ssh
    ```

#### **b. Regular Security Updates**

1. **Install `unattended-upgrades` to automate updates:**

    ```shell
    sudo apt-get install unattended-upgrades
    sudo dpkg-reconfigure --priority=low unattended-upgrades
    ```

2. **Ensure all packages are kept up to date:**

    ```shell
    sudo apt-get update && sudo apt-get upgrade -y
    ```

#### **c. Isolate the Validator Node**

- Run the validator on a dedicated server or virtual machine to avoid sharing the host with other services.
- Use virtualization (e.g., KVM, Proxmox) or containers (e.g., Docker) to separate services.


### **3. Backups and Recovery Planning**

#### **a. Regular Backups**

1. **Identify critical files to back up:**
   - `priv_validator_key.json`
   - `node_key.json`
   - `config.toml`

2. **Use `rsync` or `scp` to copy files to a secure backup server:**

    ```shell
    rsync -avz ~/.neutaro/config/ user@backupserver:/path/to/backup/
    ```

3. **Set up a cron job to automate backups:**

   - **Edit crontab:**

    ```shell
    crontab -e
    ```

   - **Add the following line for daily backups at midnight:**

    ```shell
    0 0 * * * rsync -avz ~/.neutaro/config/ user@backupserver:/path/to/backup/
    ```

#### **b. Quick Recovery Plan**

- Prepare a backup server with the necessary environment.
- Test restoring backups regularly to ensure they work as expected.
- Document steps for quick node recovery in case of a compromise.


### **4. Monitoring and Logging**

#### **a. Monitor for Unauthorized Access**

1. **Install Fail2Ban to monitor and ban suspicious activity:**

    ```shell
    sudo apt-get install fail2ban
    sudo systemctl enable fail2ban
    sudo systemctl start fail2ban
    ```

2. **Configure Fail2Ban by editing `/etc/fail2ban/jail.local`:**

    ```shell
    [sshd]
    enabled = true
    port = ssh
    filter = sshd
    logpath = /var/log/auth.log
    maxretry = 5
    ```

3. **Restart Fail2Ban to apply changes:**

    ```shell
    sudo systemctl restart fail2ban
    ```

#### **b. Audit Logs Regularly**

1. **Use `journalctl` or `grep` to search for suspicious activity:**

    ```shell
    sudo journalctl -u neutaro | grep 'error'
    ```

2. **Ensure logging is enabled in `config.toml`:**

   - Set appropriate logging levels like `info` or `debug` to get the required details.


### **5. Educate Validators on Security Best Practices**

#### **a. Create Easy-to-Understand Guides**
- **What to Include:**
  - Key security steps, server setup, and best practices for managing nodes.
- **Resources:**
  - Consider creating step-by-step guides or video tutorials for easier understanding.

#### **b. Regular Security Check-Ups**
- **Encourage Validators to Regularly Review:**
  - Check for updates.
  - Review access logs.
  - Ensure backups are functioning correctly.


### **6. Use Two-Factor Authentication (2FA) and Strong Passwords**

#### **a. Enforce Strong Password Policies:**
- Use tools like `passwd` to enforce password complexity.

#### **b. Enable Two-Factor Authentication:**
- For SSH, use Authy or other 2FA methods.
