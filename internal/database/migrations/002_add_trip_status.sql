ALTER table trips
    add column status text not null default 'requested'
        check (
            status in (
                       'requested',
                       'accepted',
                       'in_progress',
                       'completed',
                       'cancelled'
                )
            );