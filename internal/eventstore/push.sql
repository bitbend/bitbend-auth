insert into events (tenant_id, aggregate_type, aggregate_version, aggregate_id, aggregate_sequence, event_type,
                    data, resource_owner, creator, correlation_id, causation_id, global_position, created_at)
values
%s returning global_position, created_at;
