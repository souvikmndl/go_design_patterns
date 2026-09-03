# Apache Cassandra — Staff-Level System Design Notes

> A distributed, wide-column NoSQL database. Think **Dynamo (availability + partitioning) ⊕ Bigtable (wide-column storage engine)**. Built at Facebook for inbox search; now runs Discord, Netflix, Apple, Bloomberg. The single most important thing to internalize: **everything about Cassandra is a consequence of two design bets — (1) availability over consistency, and (2) write throughput over read simplicity.** Every feature below falls out of those two bets.

---

## 0. TL;DR — Cassandra at a glance

```mermaid
mindmap
  root(("Cassandra"))
    Lineage
      Dynamo
        Consistent hashing
        Tunable consistency
        Gossip + hinted handoff
      Bigtable
        Wide-column model
        LSM-tree storage
    Core bets
      Availability over consistency (AP)
      Write throughput over read cost
    Superpowers
      Horizontal scale on commodity HW
      Masterless - no SPOF
      High write throughput
      Multi-DC replication
    Costs you pay
      No JOINs
      No ACID transactions
      Query-driven modeling
      Denormalize everything
```

**One-liner for an interview:** "Cassandra is a masterless, AP-leaning wide-column store with tunable consistency. It scales writes linearly because its storage engine is an LSM tree — writes are append-only — and it stays available because there's no leader: any node can coordinate, data is replicated N ways, and it uses hinted handoff + read repair to heal. The price is that you model per-query and denormalize, because there are no joins and no cross-partition transactions."

---

## 1. When to reach for Cassandra (and when not to)

| Reach for it when… | Avoid it when… |
|---|---|
| **Write-heavy** workloads (event logs, messages, metrics, feeds) — LSM tree eats writes | You need **strict serializable consistency** / ACID across rows |
| **Availability > consistency** — must stay up through node/rack/DC loss | You need **ad-hoc queries**, JOINs, aggregations, or don't know access patterns upfront |
| **Massive scale** with predictable, known **access patterns** | Workload is **read-heavy with random point lookups across huge datasets** and you can't model partitions well |
| **Flexible / sparse schemas** — many columns, not all populated per row | Data has **strong relational structure** with lots of cross-entity queries |
| **Multi-datacenter / geo-distributed** replication needed | Small data that fits comfortably in one Postgres box (don't pay the complexity tax) |

**CAP framing (say this out loud in an interview):** Cassandra is an **AP** system by default — under a partition it keeps serving reads/writes and reconciles later via last-write-wins. But consistency is a *dial*, not a fixed setting: with `QUORUM` reads **and** writes you buy read-your-writes-style strong consistency at the cost of availability (you need a majority of replicas up). You are choosing a point on the CAP surface *per query*, which is unusually powerful.

---

## 2. Data model

The hierarchy is `Keyspace → Table → Row → Column`. Mentally, it's one giant nested map (a `Map<PartitionKey, SortedMap<ClusteringKey, Row>>` per table).

```mermaid
flowchart TD
    KS["Keyspace: chat<br/>(owns replication strategy + UDTs)"]
    KS --> T1["Table: users"]
    KS --> T2["Table: messages"]
    T1 --> R1["Row user_id=101<br/>name=Evan, email=..., age=15"]
    T1 --> R2["Row user_id=102<br/>name=Stefan, email=..., age=25, address=..."]
    T2 --> R3["Row msg_id=5<br/>sender=Evan, text='I love cassandra'"]
    T2 --> R4["Row msg_id=6<br/>sender=Evan, text='Check out hellointerview.com'"]

    note1["Note: user 101 has no address, 102 does.<br/>Wide-column = columns vary per row,<br/>no NULL padding like an RDBMS."]
    R2 -.-> note1
```

Key definitions and the **why** behind each:

- **Keyspace** — the "database". Its job is to own the **replication strategy** (how many copies, in which DCs) and any user-defined types. Replication is a keyspace-level property because you want all tables in a keyspace to share a durability/availability policy.
- **Table** — rows under a schema. But schema here means *primary key structure + declared columns*; individual rows may omit columns.
- **Row** — identified by a **primary key**, stores values across columns.
- **Column** — `(name, type, value, timestamp)`. The **timestamp is the crux of conflict resolution**: when two replicas disagree on a column, **last-write-wins (LWW)** by timestamp. This is why Cassandra needs no coordination to write — every cell carries its own version.

> **Staff-level nuance the notes skip:** LWW is *per column*, not per row. Two concurrent writes to different columns of the same row both survive and merge. But two writes to the *same* column silently drop the older one — and if clocks skew across nodes, "last" can be wrong. This is why you keep writes idempotent and avoid using Cassandra as a lock or counter-of-record without care (counters are a special CRDT-backed type for exactly this reason).

Columns support scalars, collections (`list/set/map`), UDTs, and JSON — so both flat and nested data fit.

---

## 3. Primary key — the single most important design decision

```
PRIMARY KEY ( (partition key columns), clustering columns )
                └─ which node ─┘        └─ sort order on that node ─┘
```

- **Partition key** — one or more columns hashed to decide **which partition (and therefore which nodes) own the row**. Determines *data locality*. All rows sharing a partition key live together on the same replica set.
- **Clustering key** — zero or more columns that define the **sorted order of rows within a partition**. Determines *on-disk sort* and enables efficient range scans within a partition.

```sql
-- partition key a only
PRIMARY KEY (a)

-- partition a, clustering b (sorted ASC on disk)
PRIMARY KEY ((a), b)  WITH CLUSTERING ORDER BY (b ASC)

-- COMPOSITE partition key (a+b) — both needed to locate the partition
PRIMARY KEY ((a, b), c)

-- partition a, two clustering cols b then c
PRIMARY KEY ((a), b, c)   -- == PRIMARY KEY (a, b, c)
```

**Why it matters so much:** the partition key decides whether a query is a **single-partition read** (fast, one replica set) or a **scatter-gather** across the whole cluster (slow, kills you at scale). *You design the key so your hot query hits exactly one partition.* This is identical to DynamoDB's partition-key/sort-key model — the concept maps 1:1.

**The golden constraint:** a good partition is **bounded in size** (rule of thumb: keep partitions under ~100 MB / ~100k rows) and **doesn't grow forever**. Unbounded partitions are the #1 Cassandra modeling failure (see Discord below).

---

## 4. Internals

### 4.1 Partitioning — consistent hashing + virtual nodes

Naïve hashing `node = hash(key) mod N` has two problems: **(1)** changing `N` (add/remove a node) remaps almost *every* key → massive data movement; **(2)** unlucky hashing → hot nodes.

**Consistent hashing** fixes (1): hash keys onto a ring of integers; walk **clockwise** to the first node token; store there. Adding/removing a node only remaps keys between it and its neighbor — everyone else is untouched.

**Virtual nodes (vnodes)** fix (2): each physical machine owns *many* small token ranges scattered around the ring, not one big arc. This smooths load and lets beefier machines own more vnodes (heterogeneous hardware).

```mermaid
flowchart LR
    subgraph RING["Token ring — hash(key), then walk clockwise to next token"]
      direction LR
      t1["t1"] --> t2["t2"] --> t3["t3"] --> t4["t4"] --> t5["t5"] --> t6["t6"] --> t7["t7"] --> t8["t8"] --> t1
    end
    classDef nodeA fill:#7c3aed,stroke:#fff,color:#fff;
    classDef nodeB fill:#0ea5e9,stroke:#fff,color:#fff;
    classDef nodeC fill:#f59e0b,stroke:#fff,color:#111;
    class t1,t4,t7 nodeA
    class t2,t5,t8 nodeB
    class t3,t6 nodeC
    L["Color = physical node. Each physical node owns several vnodes<br/>scattered around the ring → even load + heterogeneous capacity."]
```

*Add/remove a node* → only the **adjacent** ranges move; the rest of the cluster is undisturbed. This is why Cassandra scales elastically with minimal data reshuffling.

### 4.2 Replication — where copies go

`Replication Factor (RF)` = number of copies. To place replicas, Cassandra hashes the key to its primary vnode, then **walks clockwise** picking the next distinct vnodes as additional replicas — **skipping any vnode on a physical node already in the replica set** (so one machine dying doesn't take out multiple replicas).

```mermaid
flowchart LR
    subgraph RING["RF=3 placement"]
      direction LR
      n1 --> n2 --> n3 --> n4 --> n5 --> n6 --> n7 --> n8 --> n1
    end
    classDef hit fill:#22c55e,stroke:#fff,color:#111;
    classDef rep fill:#0ea5e9,stroke:#fff,color:#fff;
    class n2 hit
    class n3,n4 rep
    S["1 - key hashes to n2 (primary)<br/>2 - walk clockwise → replicate to n3, n4<br/>(skipping any vnode on a physical node already chosen)"]
```

**Two strategies:**
- **`NetworkTopologyStrategy`** (production default) — **rack- and DC-aware**. Places replicas across distinct racks and datacenters so a rack or DC outage can't take out a quorum. This is how you survive real-world failure domains.
- **`SimpleStrategy`** — plain clockwise walk, topology-blind. Dev/test only.

```sql
ALTER KEYSPACE ks WITH REPLICATION =
  {'class':'SimpleStrategy','replication_factor':3};

ALTER KEYSPACE ks WITH REPLICATION =
  {'class':'NetworkTopologyStrategy','dc1':3,'dc2':2};  -- 3 copies in dc1, 2 in dc2
```

### 4.3 Tunable consistency — the CAP dial

Cassandra has **no ACID transactions** across rows — only atomic + isolated writes to a single partition. What it *does* give you is a **consistency level (CL)** per read and per write: how many replicas must ack before the operation is considered successful.

Common levels: `ONE`, `TWO`, `QUORUM` (majority = `⌊RF/2⌋ + 1`), `LOCAL_QUORUM` (majority within the local DC — the production workhorse), `EACH_QUORUM`, `ALL`.

**The one formula to memorize:** strong consistency (a read is guaranteed to see the latest committed write) holds when

$$ R + W > RF $$

where `R` = read CL replica count, `W` = write CL replica count. If read-set and write-set must overlap by ≥1 replica, the read is guaranteed to touch a node that saw the write.

```mermaid
flowchart TD
    W["WRITE at QUORUM<br/>acked by n1 + n3  (2 of 3)"]
    R["READ at QUORUM<br/>reads n2 + n3  (2 of 3)"]
    O["n3 is in BOTH sets<br/>→ read is guaranteed to see the write"]
    W --> O
    R --> O
    note["RF=3 → QUORUM=2.  R(2)+W(2)=4 > RF(3) ✓<br/>Coordinator returns the cell with the newest timestamp."]
    O -.-> note
```

**Tuning table (RF=3):**

| Write CL | Read CL | R+W>RF? | Behavior |
|---|---|---|---|
| `ONE` | `ONE` | 2 > 3 ✗ | Fastest, most available, **eventually** consistent. Fine for the Ticketmaster browse UI. |
| `QUORUM` | `QUORUM` | 4 > 3 ✓ | **Strong** consistency, tolerates 1 replica down. Default "safe" choice. |
| `ALL` | `ONE` | 4 > 3 ✓ | Strong but writes fail if *any* replica is down — low write availability. |
| `LOCAL_QUORUM` | `LOCAL_QUORUM` | ✓ (within DC) | Strong within a DC, avoids cross-DC latency. Multi-DC production default. |

> **Curveball detail:** even at `QUORUM`, Cassandra is **not linearizable** for read-modify-write. Two clients doing "read balance, then write balance-10" can both read the same value and clobber each other — `R+W>RF` prevents *stale reads*, not *lost updates*. For compare-and-set you need **Lightweight Transactions** (§4.8).

Regardless of CL, background repair drives the cluster toward **eventual consistency** — all replicas converge given enough time.

### 4.4 Query routing — any node is a coordinator

There's no master. A client connects to *any* node, which becomes the **coordinator** for that request. Because every node knows the ring topology (via **gossip**), the replication strategy, and can run the hash itself, the coordinator knows exactly which replicas own the key. It fans the request out to them, enforces the CL, and returns the answer.

```mermaid
flowchart LR
    C["Client"] -->|"1 - query"| Coord["n2 (coordinator)"]
    Coord -->|"2 - hash key, find replicas"| Coord
    Coord -->|"3 - route to replicas"| R1["n7"]
    Coord -->|"3 - route to replicas"| R2["n8"]
    R1 --> Coord
    R2 --> Coord
    Coord -->|"4 - reconcile by timestamp, honor CL"| C
```

This masterless design is *why* Cassandra has no single point of failure and scales reads/writes by just adding nodes.

### 4.5 Storage engine — the LSM tree (why writes are cheap)

This is Cassandra's beating heart and the source of its write throughput. Most databases index with a **B-tree** (great for reads, but writes do random in-place disk I/O). Cassandra uses a **Log-Structured Merge (LSM) tree**: every insert/update/delete is an **append**, never an in-place edit. A row's current state is *derived* from the ordered sequence of its writes. Deletes are just special "tombstone" writes.

Three components:
- **Commit log** — an append-only **write-ahead log** on disk for durability.
- **Memtable** — an in-memory sorted structure (sorted by primary key), holds recent writes.
- **SSTable** ("Sorted String Table") — an **immutable** sorted file on disk, flushed from a memtable.

**Write path:**

```mermaid
flowchart LR
    W["0 - write arrives at node"] --> CL["1 - append to commit log<br/>(durability / crash recovery)"]
    W --> MT["2 - apply to memtable<br/>(in-memory, sorted)"]
    MT -->|"3 - flush when size/time threshold hit"| SST["SSTable (immutable, on disk)"]
    SST -->|"4 - drop the now-redundant commit log segments"| DONE["done"]
    style CL fill:#1e293b,stroke:#94a3b8,color:#fff
    style MT fill:#334155,stroke:#94a3b8,color:#fff
    style SST fill:#0f766e,stroke:#fff,color:#fff
```

A write is durable the instant it hits the commit log + memtable — no disk seek, no read-before-write. That's the throughput win.

**Read path** (harder — data for one key may be spread across memtable + several SSTables):

```mermaid
flowchart TD
    Q["read key K"] --> M{"in memtable?"}
    M -->|yes, newest| RET["merge newest cells by timestamp → return"]
    M -->|maybe more on disk| BF["consult Bloom filter per SSTable<br/>(skip SSTables that definitely lack K)"]
    BF --> IDX["for candidate SSTables, use partition index<br/>→ byte offset of K"]
    IDX --> SEEK["read cells, newest SSTable → oldest"]
    SEEK --> RET
    RET --> RR["(optionally) trigger read repair if replicas disagree"]
```

Supporting structures the notes gloss over:
- **Bloom filter** per SSTable — probabilistic "is key *maybe* here?" filter. No false negatives, so it lets the read **skip** SSTables that definitely don't contain the key. Turns a potential scan of every SSTable into 1–2 targeted lookups.
- **Partition index / summary** — maps key → byte offset within an SSTable (like a B-tree leaf pointer), so the read seeks straight to the data.

**Compaction** — because writes are append-only, one row's history piles up across many SSTables. Compaction periodically **merge-sorts SSTables into fewer, consolidated ones**, keeping only the newest cell per column and **purging tombstones** past their grace period. It's cheap because inputs are already sorted (merge-sort streaming). Strategies matter in interviews:
- **STCS** (Size-Tiered) — default, write-optimized; compacts similarly-sized SSTables. Can cause read amplification + space spikes.
- **LCS** (Leveled) — read-optimized; guarantees a key lives in ≤1 SSTable per level. Great for read-heavy, update-heavy tables; higher write amplification.
- **TWCS** (Time-Windowed) — for **time-series/TTL** data; compacts within time windows and drops whole expired SSTables cheaply. This is the one for metrics/events.

```mermaid
flowchart LR
    A["SSTable1 (K:v1)"] --> C["Compaction<br/>(merge-sort, keep newest,<br/>drop tombstones past gc_grace)"]
    B["SSTable2 (K:v2, DELETE col)"] --> C
    D["SSTable3 (K:v3)"] --> C
    C --> E["SSTable_new (K: consolidated state)"]
```

**Tombstones — the classic Cassandra footgun.** A delete writes a tombstone marker; the data isn't physically gone until compaction runs *after* `gc_grace_seconds` (default 10 days — the window ensures a deleted value doesn't "resurrect" from a replica that missed the delete, i.e. **zombie data**). Consequences: (1) a partition with lots of deletes/expiring rows accumulates tombstones that reads must scan through → latency and even OOM; (2) modeling **queues** in Cassandra is an anti-pattern precisely because you delete constantly and drown in tombstones.

### 4.6 Gossip — cluster membership

Nodes disseminate cluster state (who's alive, schema versions, load) **peer-to-peer** via **gossip**: each node periodically picks a few peers and exchanges state. Every node tracks, per known node, a **generation** (boot timestamp) and a **version** (a monotonically increasing logical clock, ~1/sec). Together these act as a **vector clock**, letting a node discard stale gossip.

```mermaid
flowchart LR
    subgraph C["Gossip (P2P, biased toward seeds)"]
      A["nodeA"] <--> B["nodeB"]
      B <--> S(("SEED"))
      A <--> S
      D["nodeD"] <--> S
      D <--> B
    end
    N["Seed nodes = always-discoverable gossip hotspots.<br/>They prevent the cluster from splitting into<br/>sub-clusters that never hear about each other."]
    S -.-> N
    style S fill:#f59e0b,stroke:#fff,color:#111
```

**Seed nodes** are designated bootstrap/hotspot nodes (found via service discovery) that gossip is biased toward, guaranteeing information reaches the whole cluster. Universal topology knowledge is *why any node can coordinate any request*.

### 4.7 Fault tolerance — detect, route around, heal

- **Failure detection — Phi Accrual Failure Detector.** Instead of a binary "is it dead?" timeout, each node computes a **suspicion level (Φ)** from the *statistical distribution of recent heartbeat inter-arrival times*. Higher Φ = more confident it's down. This adapts to network conditions (a slow link raises the threshold instead of false-positiving). When Φ crosses a threshold the node is **convicted** and taken out of write routing — but **kept in the ring** until an admin changes topology, so transient blips/restarts don't trigger expensive rebalancing.

- **Hinted handoff** — if a replica is down when a write arrives, the coordinator stores a **hint** (the write, plus "this belongs to n8") locally and proceeds. When n8 recovers, the hint is replayed to it. Hints have a **short TTL** (default 3h) — they're a *short-term* patch, not a durability guarantee.

```mermaid
sequenceDiagram
    participant Client
    participant n2 as n2 (coordinator)
    participant n7
    participant n8 as n8 (offline)
    Client->>n2: 1. write (RF=3 → n2,n7,n8)
    n2->>n7: 2. replicate ✓
    n2-xn8: 2. replicate ✗ (offline)
    n2->>n2: 3. store HINT for n8
    Note over n2: gossip later reports n8 alive
    n2->>n8: 4. hinted handoff (replay stored write)
    Note over n8: n8 now consistent
```

- **Read repair** — on a `QUORUM`+ read, if replicas return different timestamps, the coordinator returns the newest value **and** pushes the fresh value back to the stale replica(s) inline. Passive healing on the read path.
- **Anti-entropy repair (`nodetool repair`)** — periodic background reconciliation using **Merkle trees**: replicas exchange hash trees of their data ranges, diff them cheaply, and stream only the mismatched ranges. This is the *authoritative* mechanism that guarantees eventual consistency; hints and read-repair are best-effort. **You must run repair within `gc_grace_seconds`** or deleted data can resurrect.

### 4.8 Lightweight Transactions (the thing "no transactions" hides)

The notes say Cassandra has no transactions — mostly true, but there's an important exception. **Lightweight Transactions (LWT)** provide **linearizable compare-and-set on a single partition** via a **Paxos** consensus round:

```sql
UPDATE seats SET status='held' WHERE event_id=1 AND seat_id=42 IF status='available';
INSERT INTO users (email, ...) VALUES (...) IF NOT EXISTS;
```

`IF`/`IF NOT EXISTS` triggers a 4-round-trip Paxos ballot (`SERIAL` consistency) — **much slower** than a normal write, so use it only where you truly need CAS (unique constraints, seat holds, idempotency guards), never on the hot path. This is the honest answer when an interviewer asks "how would Cassandra prevent double-booking a seat?"

---

## 5. Data modeling — query-driven, not entity-driven

Relational modeling is **normalized + entity-relationship-driven**: one copy of each entity, joins to combine. Cassandra flips this: **no joins, no foreign keys, single-table queries only** → you model **per access pattern** and **denormalize (duplicate) data** across tables so every query is one partition read.

The four questions to ask for every table:
1. **Partition key** — what determines which node owns this data? (→ makes the hot query single-partition)
2. **Partition size** — worst-case rows/bytes; can it grow unbounded? (→ if yes, add a bucketing dimension)
3. **Clustering key** — what sort order does the query need?
4. **Denormalization** — what data must be duplicated so this query needs no join/aggregation?

### 5.1 Example — Discord messages (the bucketing lesson)

Access pattern: fetch **recent** messages for a **channel**, newest-first, scroll a bit.

**v1:**
```sql
CREATE TABLE messages (
  channel_id bigint, message_id bigint,
  author_id bigint, content text,
  PRIMARY KEY (channel_id, message_id)
) WITH CLUSTERING ORDER BY (message_id DESC);
```
- Partition key `channel_id` → one channel's messages live in **one partition** → single-partition read, no scatter-gather.
- Clustering `message_id DESC` → newest first for free.
- **Why `message_id` (a Snowflake) not `created_at`?** Snowflake IDs are **chronologically sortable UUIDs** — no collision (a timestamp, even at ms granularity, can collide and break primary-key uniqueness).

**Problem:** busy channels → **unbounded, ever-growing partitions** → Cassandra chokes on large partitions.

**v2 — add a `bucket`** (10-day window from a fixed `DISCORD_EPOCH`) into the **partition key**:
```sql
CREATE TABLE messages (
  channel_id bigint, bucket int, message_id bigint,
  author_id bigint, content text,
  PRIMARY KEY ((channel_id, bucket), message_id)
) WITH CLUSTERING ORDER BY (message_id DESC);
```
This **bounds** partition size (even a firehose channel fits 10 days) and stops monotonic growth (new bucket over time). Recent reads still usually hit one partition; only bucket-boundary crossings or inactive channels need a second lookup.

```mermaid
erDiagram
    messages {
        bigint channel_id "PARTITION KEY"
        int    bucket      "PARTITION KEY (10-day window)"
        bigint message_id  "CLUSTERING KEY (Snowflake, DESC)"
        bigint author_id   "denormalized"
        text   content     ""
    }
```

**Lesson:** the schema is dictated by the *access pattern* + the *partition-size constraint*. This is query-driven modeling in one example.

### 5.2 Example — Ticketmaster browse UI (denormalization)

Access pattern: browse one event's seats; UI shows a **venue map of sections** (section-level totals) then **individual seats** on click. Weak consistency is fine — availability changes live; the *purchase* flow re-checks a consistent store.

**v1 — everything keyed by event:**
```sql
CREATE TABLE tickets (
  event_id bigint, seat_id bigint, price bigint,
  PRIMARY KEY (event_id, seat_id)
);
```
Problem: a 10k+ seat event = one big partition, and every browse hits it; computing section totals means aggregating on read (expensive, repeated).

**v2 — push `section_id` into the partition key** (mirrors the UX: users drill into a section):
```sql
CREATE TABLE tickets (
  event_id bigint, section_id bigint, seat_id bigint, price bigint,
  PRIMARY KEY ((event_id, section_id), seat_id)
);
```
Now each **section is its own partition** → event spreads across nodes, each partition is small, and the "show me this section's seats" query is single-partition.

**v3 — denormalize a summary table** for the top-level venue view (so you never aggregate on read):
```sql
CREATE TABLE event_sections (
  event_id bigint, section_id bigint,
  num_tickets bigint, price_floor bigint,
  PRIMARY KEY (event_id, section_id)
);
```
Partitioned by `event_id`; events have < ~100 sections, so the whole venue map is one small single-partition read. Totals don't need to be exact (Ticketmaster shows "100+") → eventual consistency is acceptable, which is *why* denormalizing a fuzzy count is safe here.

```mermaid
erDiagram
    tickets {
        bigint event_id   "PARTITION KEY"
        bigint section_id "PARTITION KEY"
        bigint seat_id    "CLUSTERING KEY"
        bigint price      ""
    }
    event_sections {
        bigint event_id    "PARTITION KEY"
        bigint section_id  "CLUSTERING KEY"
        bigint num_tickets "denormalized aggregate"
        bigint price_floor "denormalized aggregate"
    }
```

**Lesson:** UX/access patterns drive the partition key; when a query would force an aggregation or a join, **precompute it into a denormalized table** instead.

---

## 6. Advanced features (know they exist)

- **Storage-Attached Indexes (SAI)** — modern **global secondary indexes** on non-key columns. Slower than partition-key queries but decent; let you serve *low-frequency* query patterns **without** building a whole denormalized table. Use for the occasional "find by email" that doesn't justify its own table.
- **Materialized Views** — Cassandra maintains a denormalized view table off a base table automatically, so you don't hand-write multi-table writes in your app. Convenient, but historically had consistency edge cases — know the concept, use with care.
- **Search integration** — bolt on Elasticsearch/Solr (e.g. via plugins/Lucene indexes) for full-text/complex search that Cassandra's model can't serve natively.

---

## 7. Trade-off cheat sheet

| Lever | Choice A | Choice B | Rule of thumb |
|---|---|---|---|
| **Consistency** | `ONE` (fast, available, eventual) | `QUORUM`/`LOCAL_QUORUM` (strong, `R+W>RF`) | Default to `LOCAL_QUORUM`; drop to `ONE` for browse/analytics-ish paths |
| **Storage engine** | LSM tree (write-optimized) | (B-tree — what you *didn't* pick) | Cassandra = cheap writes, costlier reads; mitigate reads with bloom filters + compaction |
| **Compaction** | STCS (write-heavy) | LCS (read-heavy) / TWCS (time-series) | Match to the table's read/write ratio and TTL behavior |
| **Modeling** | Normalize | **Denormalize per query** | Duplicate data; one table per access pattern |
| **Partition key** | Coarse (few big partitions) | Fine + **bucketed** (many bounded partitions) | Keep < ~100MB / 100k rows; add a time/bucket dimension to bound growth |
| **Replica placement** | SimpleStrategy | **NetworkTopologyStrategy** | Always NTS in prod (rack/DC aware) |
| **CAS / uniqueness** | Normal write (LWW) | **LWT (Paxos, `IF`)** | Use LWT only where you truly need compare-and-set; it's slow |
| **Healing** | Hinted handoff + read repair (best-effort) | `nodetool repair` (authoritative, Merkle) | Run repair within `gc_grace_seconds` to prevent zombie data |

---

## 8. Interview positioning

**Say yes to Cassandra when:** availability > consistency, very high **write throughput**, huge scale, **known access patterns**, flexible/sparse or wide schemas, multi-DC. Canonical fits: **messaging/chat, activity feeds, time-series/metrics, event logging, IoT ingestion, notifications**.

**Say no when:** you need **strict consistency/ACID**, **JOINs / ad-hoc aggregations / analytical queries**, or you **don't yet know your access patterns**. Reach for Postgres/Spanner/DynamoDB-with-transactions instead, or pair Cassandra with a search/OLAP system for the queries it can't serve.

---

## 9. Mock interview Q&A — staff-level curveballs

**Q1. "You write at `QUORUM` and read at `QUORUM` with RF=3. A client reads its own recent write and gets an old value. How?"**
It shouldn't for a *committed* `QUORUM` write, because `R+W=4>3` forces overlap. The usual culprits: the write **hadn't returned success yet** (partially applied, coordinator not yet acked — you read a replica that hadn't gotten it); or the two ops went to **different datacenters** and you used `LOCAL_QUORUM` on each with async cross-DC replication (no overlap guarantee across DCs); or **clock skew** made an older write win LWW. *Design tweak:* keep clients pinned to one DC for read-after-write, use `LOCAL_QUORUM` consistently, and rely on NTP/monotonic write timestamps.

**Q2. "RF=3, write CL=`QUORUM`, one replica is down. Does the write succeed? What about that third replica?"**
Yes — `QUORUM`=2, two replicas ack, so it succeeds and stays available. For the down replica the coordinator stores a **hint** and replays it on recovery (hinted handoff). If the node is down longer than the hint TTL (3h), the hint is dropped and the replica is reconciled later by **read repair** or `nodetool repair`. *Trade-off surfaced:* availability is preserved but that replica is temporarily stale — durability of the "third copy" leans on repair, not hints.

**Q3. "Your read latency p99 exploded on a table that gets lots of deletes/TTLs. Why, and what do you change?"**
**Tombstones.** Deletes/expirations write markers that live until compaction after `gc_grace_seconds`; reads must scan past them, and a partition full of tombstones tanks latency (or trips the tombstone-per-query limit / OOMs). *Fixes:* switch to **TWCS** so whole expired SSTables drop cheaply; model so you **drop entire partitions** instead of deleting rows (e.g., time-bucketed partitions you can let expire); lower `gc_grace_seconds` **only if** you can guarantee repair within the window; stop using Cassandra as a queue.

**Q4. "How do you prevent double-booking the same seat at checkout in Cassandra?"**
Normal writes are LWW and will silently lose the race. Use a **Lightweight Transaction**: `UPDATE seats SET status='held', holder=? WHERE event_id=? AND seat_id=? IF status='available'` — Paxos gives linearizable CAS on that single partition. It's ~4 round-trips at `SERIAL`, so confine it to the checkout path; keep the browse path on cheap `ONE`/`LOCAL_QUORUM` reads. *Design tweak:* seat identity must be in one partition for the LWT to be single-partition.

**Q5. "Why not just add a secondary index instead of denormalizing into another table?"**
Native secondary indexes are **local per node** → a query by an indexed non-key column becomes a **scatter-gather** across all nodes (fine at low scale, terrible at high cardinality/large clusters). **SAI** improves this and is worth it for *low-frequency* patterns, but for a **hot** query you still want a **denormalized table** partitioned by that query's key so it's a single-partition read. *Rule:* denormalize the frequent access patterns; index (SAI) the rare ones.

**Q6. "Design a partition key for per-user notification history that scales for both a normal user and a celebrity with millions of events."**
Start `PRIMARY KEY (user_id, event_id DESC)` for newest-first single-partition reads — but a celebrity is an **unbounded hot partition**. Add a **time bucket**: `PRIMARY KEY ((user_id, bucket), event_id DESC)` (bucket = day/week). Recent reads hit the current bucket; older reads page across buckets. *Trade-off:* cross-bucket reads for deep history and a bit more app logic, in exchange for bounded partitions and no hotspot. Same pattern as Discord.

**Q7. "Cassandra is 'masterless,' yet a coordinator picks replicas and enforces CL. Where's the single point of failure?"**
There isn't one for the data path: **any** node can coordinate (gossip gives everyone full topology), replicas are chosen by deterministic consistent-hashing, and failure detection routes around dead nodes. The only special roles are **seed nodes**, and they're only special for *bootstrapping/gossip convergence* — losing a seed doesn't stop reads/writes; you just want seeds discoverable so new nodes can join and gossip stays connected.

**Q8. "When would you deliberately choose Cassandra over DynamoDB, given the models are nearly identical?"**
Both share the partition/clustering-key model and AP leanings. Choose Cassandra for **no-vendor-lock-in / multi-cloud or on-prem**, **full control of consistency and compaction/tuning**, and **cheaper multi-DC at massive scale**; choose Dynamo for **zero-ops serverless**, tight AWS integration, and built-in transactions. The honest staff answer: it's usually an ops/cost/lock-in decision, not a data-model one.
