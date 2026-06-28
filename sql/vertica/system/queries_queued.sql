/* Refer to the view definition for details on the underlying DC tables */
select * from vs_system_views
where view_name = 'resource_acquisitions';

with base as (
	select 
		pool_name,
		node_name,
		transaction_id,
		acquisition_timestamp,
		queue_entry_timestamp,
		release_timestamp,
		(extract(epoch from(acquisition_timestamp - queue_entry_timestamp)) * 1000)::int queue_wait_ms,
		(extract(epoch from(release_timestamp - acquisition_timestamp)) * 1000)::int duration_ms_calc,
		duration_ms,
		thread_count,
		open_file_handle_count,
		memory_inuse_kb
	from resource_acquisitions
	where null is null
	and acquisition_timestamp > now() - interval '1 hour'
), per_trx as (
	select 
		pool_name,
		transaction_id,
		min(queue_entry_timestamp) queue_entry_timestamp,
		avg(queue_wait_ms)::int queue_wait_ms,
		avg(duration_ms_calc)::int duration_ms_calc,
		avg(duration_ms)::int duration_ms,
		avg(thread_count)::int thread_count,
		avg(open_file_handle_count)::int open_file_handle_count,
		avg(memory_inuse_kb)::int memory_inuse_kb
	from base
	group by 1,2
)
select 
	pool_name,
	max(queue_wait_ms) queue_wait_ms_max,
	sum(queue_wait_ms) queue_wait_ms_total,
	min(queue_entry_timestamp) s,
	max(queue_entry_timestamp) e
from per_trx
group by 1
order by queue_wait_ms_max desc;