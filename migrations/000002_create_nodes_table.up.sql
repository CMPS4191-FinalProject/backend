CREATE TYPE node_status AS ENUM ('ONLINE', 'OFFLINE', 'ERROR');

CREATE TABLE "Nodes" (
  "device_id" serial PRIMARY KEY,
  "status" node_status NOT NULL DEFAULT 'OFFLINE'
);
