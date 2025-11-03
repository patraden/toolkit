-- resource pool status sorted by memory consumption
select
    pool_name,
    sysdate as current_time,
    node_name,
    query_budget_kb,
    memory_size_kb as node_memory_size_kb,
    execution_parallelism as execution_parallelism,
    memory_inuse_kb as node_memory_inuse_kb,
    general_memory_borrowed_kb as node_general_memory_borrowed_kb,
    cpu_affinity_set,
    cpu_affinity_mode,
    cpu_affinity_mask,
    running_query_count as node_running_query_count,
    (sum(memory_inuse_kb) over(partition by pool_name, sysdate)/1024/1024)::numeric(10,2) as memory_inuse_gb,
    (sum(general_memory_borrowed_kb) over(partition by pool_name, sysdate)/1024/1024)::numeric(10,2) as general_memory_borrowed_gb,
    sum(running_query_count) over(partition by pool_name, sysdate) as running_query_count
from resource_pool_status
where not is_internal
order by memory_inuse_kb desc, pool_name, sysdate, node_name;


-- threads in vertica processes
select
    date_trunc('minute', start_time) start_time,
    process,
    max(open_files_max) open_files_max,
    max(threads_max) threads_max,
    sum(files_open_start_value) total_files_open,
    avg(files_open_start_value) avg_files_open,
    sum(thread_count_start_value) total_thread_count,
    avg(thread_count_start_value) avg_thread_count,
    sum(detached_threads_start_value) total_detached_threads,
    avg(detached_threads_start_value) avg_detached_threads,
    sum(thread_stacks_avail_start_value) total_thread_stacks_avail,
    avg(thread_stacks_avail_start_value) avg_thread_stacks_avail
from dc_process_info_by_minute
where time > now() - interval '15 minutes'
group by 1,2
order by 1 desc,2;

-- threads starvation. Useful for debugging and performance tuning.
select
    pool_name,
    resource_type,
    reason,
    sum(rejection_count) as rejection_count,
    min(first_rejected_timestamp) first_rejected_timestamp,
    max(last_rejected_timestamp) last_rejected_timestamp,
    max(last_rejected_timestamp) last_rejected_timestamp,
    max(last_rejected_value) last_rejected_value
from v_monitor.resource_rejections
where lower(resource_type) = 'threads'
group by
    pool_name,
    resource_type,
    reason
order by rejection_count desc;

-- memory starvation. Useful for debugging and performance tuning.
select
    pool_name,
    resource_type,
    reason,
    sum(rejection_count) as rejection_count,
    min(first_rejected_timestamp) first_rejected_timestamp,
    max(last_rejected_timestamp) last_rejected_timestamp,
    max(last_rejected_timestamp) last_rejected_timestamp,
    (max(last_rejected_value)/1024/1024)::numeric(10,2) last_rejected_value_gb
from v_monitor.resource_rejections
where lower(resource_type) = 'memory(kb)'
group by
    pool_name,
    resource_type,
    reason
order by rejection_count desc;