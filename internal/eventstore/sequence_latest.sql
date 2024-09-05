with existing as (
    %s
    )
select events.tenant_id
     , events.owner
     , events.aggregate_type
     , events.aggregate_id
     , events.aggregate_sequence
from events
         join
     existing
     on
         events.tenant_id = existing.tenant_id
             and events.aggregate_type = existing.aggregate_type
             and events.aggregate_id = existing.aggregate_id
             and events.aggregate_sequence = existing.aggregate_sequence
    for update;