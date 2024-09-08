
echo "Stopping and disabling the Neutaro service..."
sudo systemctl stop Neutaro
sudo systemctl disable Neutaro

echo "Removing Neutaro service file..."
sudo rm -f /etc/systemd/system/Neutaro.service

echo "Reloading systemd daemon..."
sudo systemctl daemon-reload

echo "Removing Neutaro and Cosmovisor binaries and configurations..."
sudo rm -rf $HOME/.Neutaro
sudo rm -rf $HOME/Neutaro
sudo rm -rf /usr/local/bin/Neutaro
sudo rm -rf $HOME/go/bin/cosmovisor
sudo rm -rf $HOME/.bash_profile

echo "Cleaning up Go installation..."
sudo rm -rf /usr/local/go

echo "Cleaning up any residual files..."
sudo rm -rf /root/.Neutaro
sudo rm -rf $HOME/.cache/go-build
sudo rm -rf /var/log/Neutaro*

echo "Neutaro setup has been completely removed from your system."
