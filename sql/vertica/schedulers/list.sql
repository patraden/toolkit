select
    mbs.source topic,
    mbs.cluster kafka,
    mbs.partitions topic_partitions,
    mb.id,
    mb.microbatch,
    mb.enabled,
    mb.max_parallelism,
    mb.target_columns,
    mbt.target_schema,
    mbt.target_table,
    ls.parser,
    ls.parser_parameters,
    ls.load_method,
    ls.id
from :schema.stream_microbatches mb
join :schema.stream_load_specs ls
    on ls.id = mb.load_spec
join :schema.stream_microbatch_source_map mbm
    on mb.id = mbm.microbatch
join :schema.stream_sources mbs
    on mbs.id = mbm.source
join :schema.stream_targets mbt
    on mb.target = mbt.id;