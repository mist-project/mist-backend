# — add 1Password repo & key, install CLI (must be root)
echo "https://downloads.1password.com/linux/alpinelinux/stable/" >> /etc/apk/repositories
wget https://downloads.1password.com/linux/keys/alpinelinux/support@1password.com-61ddfc31.rsa.pub \
     -P /etc/apk/keys
apk update && apk add 1password-cli

# verify install
op --version

# — fetch credentials from 1Password
TENANT_USERNAME=$(op item get $TENANT_ACCOUNT_ID --vault "$OP_VAULT" --fields TENANT_USERNAME --reveal)
TENANT_PRIVATE_SSH_KEY=$(op item get $TENANT_ACCOUNT_ID --vault "$OP_VAULT" --fields TENANT_PRIVATE_SSH_KEY)
TENANT_ADDRESS=$(op item get $TENANT_ACCOUNT_ID --vault "$OP_VAULT" --fields TENANT_ADDRESS)

# append to tmpenvs and load it
KEY_FILE=$(mktemp)
touch .tmpenvs
echo "TENANT_USERNAME=$TENANT_USERNAME" >> .tmpenvs
echo "TENANT_PRIVATE_SSH_KEY_FILE=$KEY_FILE" >> .tmpenvs
echo "TENANT_ADDRESS=$TENANT_ADDRESS" >> .tmpenvs

chmod 600 "$KEY_FILE"
echo "$TENANT_PRIVATE_SSH_KEY" > "$KEY_FILE"