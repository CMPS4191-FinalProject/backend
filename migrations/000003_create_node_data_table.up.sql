CREATE TABLE "NodeData" (
  "id" serial PRIMARY KEY,
  "user_id" integer NOT NULL,
  "device_id" integer NOT NULL,
  "moisture_content" float,
  "timestamp" timestamp DEFAULT CURRENT_TIMESTAMP
);
