useradd -m -s /bin/bash -p $(openssl passwd -1 "12345") foo && usermod -aG sudo foo

echo 'PasswordAuthentication yes' >> /etc/ssh/sshd_config.d/60-cloudimg-settings.conf && systemctl restart ssh



mkdir -p ~foo/.ssh && echo 'ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOLgt6ChgzYPLQC/6V5z/d+vIUf1pp4LP0qhLMPnjhFp blago@blago-l450' >> ~foo/.ssh/authorized_keys && chown -R foo:foo ~foo/.ssh && chmod 700 ~foo/.ssh && chmod 600 ~foo/.ssh/authorized_keys
