CREATE TABLE users (
  id text PRIMARY KEY
);

drop table money_transfers;

CREATE TABLE transfers (
  -- this may mean that
  -- a) <debtor> has borrowed <amount> of <currency> from <creditor> (debts)
  -- b) <creditor> has paid <amount> of <currency> for a good for <debtor> (debts)
  -- c) <creditor> has given <debtor> money in repayment of a previous debt (settlements)
  -- d) <creditor> is charging interest on <debtor> (interest charges)
  -- e) <debtor> will owe <creditor> if <creditor> presents some condition (conditional debt)

  id text PRIMARY KEY, -- unique identifier
  debtor text NOT NULL, -- username@server.name
  creditor text NOT NULL, -- username@server.name
  st_there timestamp, -- server date as received from peer
  st_here timestamp NOT NULL DEFAULT now(), -- server date
  actual_date datetime,
  amount int NOT NULL, -- in the smallest enumerable, 5 USD is actually 500 here
  currency text NOT NULL, -- 
  description text NOT NULL,
  next jsonb,
  signature text NOT NULL, -- of creditor:debtor:t:amount:currency[:next.creditor:next.amount:next.currency:next.onion]

  CHECK (creditor != debtor)
);

CREATE TABLE acks (
  creditor text NOT NULL,
  transfer_id text UNIQUE NOT NULL,
  signature text NOT NULL, -- of creditor:transfer_id:t
  st_there timestamp, -- server date as received from peer
  st_here timestamp NOT NULL DEFAULT now() -- server date
)

CREATE TABLE trustlines (
  truster text NOT NULL, -- someone@some.where
  trusted text NOT NULL, -- someone@some.where
  currency text NOT NULL,
  amount text NOT NULL,

  UNIQUE(truster, trusted, currency)
);
