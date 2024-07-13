CREATE TABLE gophermart.users_orders (
    user_id INT REFERENCES gophermart.users(id) ON UPDATE CASCADE ON DELETE CASCADE ,
    order_id INT REFERENCES gophermart.orders(id) ON UPDATE CASCADE ON DELETE CASCADE ,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT action_tag_pkey PRIMARY KEY (user_id, order_id)  -- explicit pk
);
