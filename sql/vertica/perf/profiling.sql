/**
https://www.vertica.com/docs/11.0.x/HTML/Content/Authoring/AdministratorsGuide/Profiling/ProfilingDatabasePerformance.htm
https://www.vertica.com/kb/System-Tables-For-Performance/Content/BestPractices/System-Tables-For-Performance.htm
*/

-- Step 1: Start a transaction and get the current trx id and statement.
START TRANSACTION;
-- current trx id and statement to be able to track your query.
SELECT current_trans_id(), current_statement();
-- will return something like: 67553995225823170, 14

-- Step 2: Run the query with "profile" keyword and get the result.
/**
Example:

profile
select
    st.ad_scheme            AS ad_scheme,
    SUM(st.prerequests)     AS prerequests,
    SUM(st.requests)        AS requests,
    SUM(st.impressions)     AS impressions,
    0                       AS clicks,
    SUM(st.conversions)     AS conversions,
    SUM(st.revenue)         AS revenue,
    SUM(st.costs)           AS costs,
    SUM(st.pub_impressions) AS pub_impressions
FROM dm_onclick.full_hourly st
join adp.adp_pst_zones_channel_format ch
    on ch.id = st.zone_id
        and ch.zone_channel_id in (1, 3)
WHERE true
  and st.date_time >= '2025-10-10 00:00:00'
  AND st.date_time <= '2025-10-22 23:59:59'
  AND st.direction_id = '1'
group by 1
order by 1;
*/

-- Step 3: Find the statement in the dc_requests_issued table.
select *
from dc_requests_issued
where transaction_id  = 67553995225823170
and request_type = 'QUERY';