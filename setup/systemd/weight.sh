[Unit]
Description=Weight Challenge API
Requires=docker.service
After=docker.service

[Service]
Restart=always
ExecStart=docker start weight-challenge-api
ExecStop=docker stop weight-challenge-api

[Install]
WantedBy=multi-user.target