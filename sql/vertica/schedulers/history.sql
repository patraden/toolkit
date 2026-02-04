select
    source_name,
    source_cluster,
    source_partition,
    start_offset,
    end_offset,
    batch_start,
    batch_end
from
(   select
        *,
        extract(milliseconds from last_batch_duration::interval second) as ld
    from :schema.stream_microbatch_history
limit :depth over (
    partition by
        microbatch_id,
        target_schema,
        target_table,
        source_name,
        source_cluster,
        source_partition
    order by epoch desc, frame_start desc)
) as lastframe
where microbatch_id = :microbatch_id
order by
    microbatch_id,
    target_schema,
    target_table,
    source_name,
    source_cluster,
    source_partition;