ALTER TABLE "NodeData"
  ADD CONSTRAINT "fk_node_data_device_id"
  FOREIGN KEY ("device_id") REFERENCES "Nodes" ("device_id");

ALTER TABLE "NodeData"
  ADD CONSTRAINT "fk_node_data_user_id"
  FOREIGN KEY ("user_id") REFERENCES "Users" ("user_id")
  ON UPDATE CASCADE
  ON DELETE SET DEFAULT;
