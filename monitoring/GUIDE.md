## Guide to setup graphana

1. Run `docker compose up --build -d` to start the Grafana server.
2. Log in with admin,admin as username and password.
3. Add a new data source:
   - Click on "Add your first data source".
   - Select
   - Choose "Prometheus" as the data source type.
   - Set the URL to `http://prometheus:9090` and all the rest default.
4. Import a dashboard:
   - Click on the "+" icon in the left sidebar.
   - Select "Import".
   - Import from json file `cookiedashboard.json`.
