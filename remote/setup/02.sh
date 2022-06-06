read -p "Enter username for mailtrap account: " MAILTRAP_USER
read -p "Enter password for mailtrap account: " MAILTRAP_PW

echo "MAILTRAP_USER='${MAILTRAP_USER}'" >> /etc/environment
echo "MAILTRAP_PW='${MAILTRAP_PW}'" >> /etc/environment
