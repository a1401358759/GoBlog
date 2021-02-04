drop procedure if exists auto_approve_revision;
create
    procedure auto_approve_revision(
        IN exists_max_revision_id int,
        IN revision_id int,
        IN publication_state_code int
    )
begin
    declare group_id varchar(256);
    declare group_name varchar(256);
    declare action_id int;
    declare done int default false;
    declare auto_approve_cursor cursor for
    select TargetGroupID, TargetGroupName, ActionID from deployment where RevisionID = exists_max_revision_id and TargetGroupID != 'D374F42A-9BE2-4163-A0FA-3C86A401B7A7';
    declare continue HANDLER for not found set done = true;
    open auto_approve_cursor;
    repeat
        fetch auto_approve_cursor into group_id, group_name, action_id;
        if done = false then
            -- 过期更新不允许审批到安装
            if not (action_id = 0 and publication_state_code = 1) then
                select count(1) into @deployment_count from deployment where RevisionID = revision_id and TargetGroupID = group_id and ActionID = action_id;
                if @deployment_count = 0 then
                    delete from deployment where RevisionID = revision_id and TargetGroupID = group_id;
                    insert into deployment (version_id, RevisionID, TargetGroupID, TargetGroupName, ActionID, DeploymentGuid, LastChangeTime, AdminName)
                    values (1, revision_id, group_id, group_name, action_id, UUID(), utc_timestamp(6), 'CMOSservice');
                end if;
            end if;
        end if;
    until done end repeat;
    close auto_approve_cursor;
    commit ;
end
