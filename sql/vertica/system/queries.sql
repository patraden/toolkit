-- top 10 queries by duration every past 12 hours. Useful for debugging and performance tuning.

with base as (
select
    i.transaction_id,
    i.request_id,
    i.statement_id,
    date_trunc('hour', i.time) as start_time_hour,
    EXTRACT(EPOCH FROM (coalesce(c.time, now()) - i.time)) AS duration_seconds,
    i.user_name,
    c.reserved_extra_memory,
    c.processed_row_count,
    c.success,
    count(*) over() qty,
    i.time as start_time,
    c.time as end_time,
    i.request,
    row_number() over(partition by date_trunc('hour', i.time) order by EXTRACT(EPOCH FROM (coalesce(c.time, now()) - i.time))  desc) rn
from dc_requests_issued i
left join dc_requests_completed c
    USING (node_name, session_id, request_id, statement_id, transaction_id)
where i.request_type = 'QUERY'
    and i.time > now() - interval '12 hour'
    and i.user_name = 'username'
)
select *
from base where rn <= 10
order by start_time_hour desc, duration_seconds desc;


-- queires in past 2 hours with resource consumptions. Useful for debugging and performance tuning.

with execution_summaries as (
    select
        transaction_id,
        node_name,
        session_id,
        request_id,
        statement_id,
        thread_count,
        (peak_memory_kb/1024/1024)::numeric(10,2) as peak_memory_gb,
        ((duration_us / 1000000))::int duration_sec,
        cpu_time_us,
        (data_bytes_read/1024/1024/1024)::numeric(10,2) as data_gb_read,
        (data_bytes_written/1024/1024/1024)::numeric(10,2) as data_gb_written,
        (network_bytes_sent/1024/1024/1024)::numeric(10,2) as network_gb_sent,
        (network_bytes_received/1024/1024/1024)::numeric(10,2) as network_gb_received,
        (bytes_spilled/1024/1024/1024)::numeric(10,2) as gb_spilled
    from dc_execution_summaries
)
select
    i.session_id,
    i.transaction_id,
    i.request_id,
    i.statement_id,
    i.time as start_time,
    c.time as end_time,
    EXTRACT(EPOCH FROM (coalesce(c.time, now()) - i.time)) AS duration_seconds,
    i.user_name,
    c.reserved_extra_memory,
    c.success,
    e.node_name,
    e.thread_count,
    e.data_gb_read,
    e.data_gb_written,
    e.duration_sec,
    e.gb_spilled,
    e.peak_memory_gb,
    e.network_gb_received,
    e.network_gb_sent
from dc_requests_issued i
left join dc_requests_completed c
    USING (node_name, session_id, request_id, statement_id, transaction_id)
left join execution_summaries e
    using (session_id, statement_id, transaction_id)
where i.request_type = 'QUERY'
    and i.user_name = 'username'
    and i.time > now() - interval '2 hours'
order by i.time desc, i.transaction_id, e.statement_id, e.node_name;