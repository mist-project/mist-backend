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
INVENTORY_FILE="ansible/inventory/hosts.ini"

# Create ansible inventory directory and temporary file
mkdir -p ansible/inventory
touch $INVENTORY_FILE
echo "[mist-service]" >> "$INVENTORY_FILE"
echo $TENANT_ADDRESS >> "$INVENTORY_FILE"

# Create PRIVATE_SSH_KEY temporary file
touch $KEY_FILE
chmod 600 "$KEY_FILE"
echo -e "$TENANT_PRIVATE_SSH_KEY" >> "$KEY_FILE"


# Create tmporary environment variables file
touch .tmpenvs
echo "export TENANT_USERNAME=$TENANT_USERNAME" >> ".tmpenvs"