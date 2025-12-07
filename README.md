# jsonsql

Go の `database/sql` で JSON / JSONB カラムを扱うための軽量ジェネリクス型です。

## インストール

```bash
go get github.com/jinford/jsonsql
```


## 比較

### SELECT して値を取り出す

<table>
<tr><th>stdlib</th><th>jsonsql</th></tr>
<tr>
<td>

```go
type UserRow struct {
    ID      int64
    Profile json.RawMessage
}

var row UserRow
err := db.QueryRowContext(ctx, query, id).
    Scan(&row.ID, &row.Profile)
if err != nil {
    return err
}

var profile Profile
if err := json.Unmarshal(row.Profile, &profile); err != nil {
    return err
}
```

</td>
<td>

```go
type UserRow struct {
    ID      int64
    Profile jsonsql.Value[Profile]
}

var row UserRow
err := db.QueryRowContext(ctx, query, id).
    Scan(&row.ID, &row.Profile)
if err != nil {
    return err
}

profile := row.Profile.Get()
```

</td>
</tr>
</table>

### INSERT する

<table>
<tr><th>stdlib</th><th>jsonsql</th></tr>
<tr>
<td>

```go
profileJSON, err := json.Marshal(profile)
if err != nil {
    return err
}

_, err = db.ExecContext(ctx,
    "INSERT INTO users (profile) VALUES ($1)",
    profileJSON,
)
```

</td>
<td>

```go
_, err := db.ExecContext(ctx,
    "INSERT INTO users (profile) VALUES ($1)",
    jsonsql.NewValue(profile),
)
```

</td>
</tr>
</table>

### NULL を扱う

<table>
<tr><th>stdlib</th><th>jsonsql</th></tr>
<tr>
<td>

```go
type UserRow struct {
    Meta json.RawMessage // NULL の可能性あり
}

var row UserRow
err := db.QueryRowContext(ctx, query, id).
    Scan(&row.Meta)

var meta map[string]any
if row.Meta != nil {
    if err := json.Unmarshal(row.Meta, &meta); err != nil {
        return err
    }
}
```

</td>
<td>

```go
type UserRow struct {
    Meta jsonsql.Nullable[map[string]any]
}

var row UserRow
err := db.QueryRowContext(ctx, query, id).
    Scan(&row.Meta)

if meta, ok := row.Meta.Get(); ok {
    // meta を使用
}
```

</td>
</tr>
</table>

## ライセンス

MIT License
