-- data collector table listing
select
    table_name,
    description,
    access_restricted
from v_monitor.data_collector
group by 1,2,3
order by 1;

-- system views
select * from vs_system_views;

-- system tables
select * from vs_system_tables;