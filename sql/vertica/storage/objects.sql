/** top tables by used space. */

select
    anchor_table_schema as schema,
    anchor_table_name as table,
    round(sum(used_bytes) / ( 1024^3 ),0 )::int as used_compressed_gb
from v_monitor.column_storage
left join tables t
    on anchor_table_schema = t.table_schema
    and anchor_table_name = t.table_name
    group by 1,2
order by used_compressed_gb desc;


/** projections storage for table */
select
    ps.projection_name,
    ps.anchor_table_schema,
    ps.anchor_table_name,
    ps.node_name,
    round(sum(ps.used_bytes)/1024/1024/1024, 5)::numeric(12,1) as gb,
    sum(ps.row_count) as rows
from projections p
join projection_storage ps
    on p.projection_id = ps.projection_id
where p.is_segmented
    and ps.anchor_table_schema = :schema
    and ps.anchor_table_name = :table
group by 1, 2, 3, 4
order by projection_name, node_name;