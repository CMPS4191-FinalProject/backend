ALTER TABLE "NodeFavorites" ADD CONSTRAINT "fk_node_favorites_user_id" FOREIGN KEY ("user_id") REFERENCES "Users" ("user_id");
ALTER TABLE "NodeFavorites" ADD CONSTRAINT "fk_node_favorites_device_id" FOREIGN KEY ("device_id") REFERENCES "Nodes" ("device_id");
