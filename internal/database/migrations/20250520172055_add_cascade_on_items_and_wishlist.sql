-- +goose Up
-- +goose StatementBegin
ALTER TABLE items DROP CONSTRAINT IF EXISTS items_wishlist_id_fkey;

ALTER TABLE items
    ADD CONSTRAINT items_wishlist_id_fkey
        FOREIGN KEY (wishlist_id) REFERENCES wishlists(id) ON DELETE CASCADE;

ALTER TABLE item_picks DROP CONSTRAINT IF EXISTS item_picks_item_id_fkey;

ALTER TABLE item_picks
    ADD CONSTRAINT item_picks_item_id_fkey
        FOREIGN KEY (item_id) REFERENCES items(id) ON DELETE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE items DROP CONSTRAINT IF EXISTS items_wishlist_id_fkey;

ALTER TABLE items
    ADD CONSTRAINT items_wishlist_id_fkey
        FOREIGN KEY (wishlist_id) REFERENCES wishlists(id);  -- without cascade

ALTER TABLE item_picks DROP CONSTRAINT IF EXISTS item_picks_item_id_fkey;

ALTER TABLE item_picks
    ADD CONSTRAINT item_picks_item_id_fkey
        FOREIGN KEY (item_id) REFERENCES items(id);  -- without cascade
-- +goose StatementEnd
