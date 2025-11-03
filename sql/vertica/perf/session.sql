-- session level isolation
SET SESSION CHARACTERISTICS AS TRANSACTION ISOLATION LEVEL SERIALIZABLE;
SHOW TRANSACTION_ISOLATION;

-- set session resource pool
SET SESSION RESOURCE POOL mypool;

-- running sessions with resources
select
    r.pool_name,
    s.node_name as initiator_node,
    s.session_id,
    r.transaction_id,
    r.statement_id,
    max(s.user_name) as user_name,
    max(substr(s.current_statement, 1, 100)) as statement_running,
    max(r.thread_count) as threads,
    max(r.open_file_handle_count) as fhandlers,
    max(r.memory_inuse_kb) as max_mem,
    count(distinct r.node_name) as nodes_count,
    min(r.queue_entry_timestamp) as entry_time,
    max(r.acquisition_timestamp - r.queue_entry_timestamp) as waiting_queue,
    max(clock_timestamp() - r.queue_entry_timestamp) as running_time
from
    v_internal.vs_resource_acquisitions r
join
    v_monitor.sessions s
    on r.transaction_id = s.transaction_id
    and r.statement_id = s.statement_id
    and length(s.current_statement) > 0
group by
    r.pool_name,
    s.node_name,
    s.session_id,
    r.transaction_id,
    r.statement_id
order by
    user_name, running_time desc, pool_name;