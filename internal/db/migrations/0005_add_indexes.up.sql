CREATE INDEX idx_heartbeats_device_received ON device_heartbeats (device_id, received_at DESC);
CREATE INDEX idx_detections_device_received ON detection_events (device_id, received_at DESC);
CREATE INDEX idx_detections_type_received ON detection_events (waste_type, received_at DESC);
CREATE INDEX idx_readings_device_bin_received ON sensor_readings (device_id, bin_type, received_at DESC);
