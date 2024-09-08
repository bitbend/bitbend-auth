insert into events (tenant_id, aggregate_type, aggregate_version, aggregate_id, aggregate_sequence, event_type, data,
                    owner, creator, correlation_id, causation_id, position, created_at)
values %s returning "position", created_at;