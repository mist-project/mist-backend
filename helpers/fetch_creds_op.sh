# — add 1Password repo & key, install CLI (must be root)
echo "https://downloads.1password.com/linux/alpinelinux/stable/" >> /etc/apk/repositories
wget https://downloads.1password.com/linux/keys/alpinelinux/support@1password.com-61ddfc31.rsa.pub \
     -P /etc/apk/keys
apk update && apk add 1password-cli && apk add coreutils

# verify install
op --version

# — fetch credentials from 1Password
TENANT_USERNAME=$(op item get $TENANT_ACCOUNT_ID --vault "$OP_VAULT" --fields TENANT_USERNAME --reveal)
TENANT_PRIVATE_SSH_KEY=$(op item get $TENANT_ACCOUNT_ID --vault "$OP_VAULT" --fields TENANT_PRIVATE_SSH_KEY | tr -d '"')
TENANT_ADDRESS=$(op item get $TENANT_ACCOUNT_ID --vault "$OP_VAULT" --fields TENANT_ADDRESS)


# Define file paths
KEY_FILE="key.pem"

# TENANT_ADDRESS_FILE="/tmp/tenant_address.json"

# Create files if they do not exist
touch .tmpenvs
touch $KEY_FILE
chmod 600 "$KEY_FILE"
# touch $TENANT_ADDRESS_FILE

# Add private keys to their corresponding files
# echo "ansible_user: $TENANT_USERNAME" > "$TENANT_USERNAME_FILE"
# echo "ansible_user: $TENANT_USERNAME" > "$TENANT_USERNAME_FILE"

# echo "ansible_ssh_private_key_file: |" > "$KEY_FILE"
echo -e "$TENANT_PRIVATE_SSH_KEY" >> "$KEY_FILE"
# truncate -s -1 $KEY_FILE
# echo "$TENANT_PRIVATE_SSH_KEY" > "$KEY_FILE"

echo "export TENANT_USERNAME=$TENANT_USERNAME" >> ".tmpenvs"
echo "export TENANT_PRIVATE_SSH_KEY_FILE=$KEY_FILE" >> ".tmpenvs"
echo "export TENANT_ADDRESS=\"$TENANT_ADDRESS,\"" >> ".tmpenvs"