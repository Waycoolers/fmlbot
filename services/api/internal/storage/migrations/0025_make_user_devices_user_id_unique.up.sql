DELETE FROM user_devices AS a
    USING user_devices AS b
WHERE a.user_id = b.user_id
  AND a.fcm_token <> b.fcm_token
  AND a.created_at < b.created_at;

alter table user_devices
    add constraint user_devices_user_id_key
        unique (user_id);
