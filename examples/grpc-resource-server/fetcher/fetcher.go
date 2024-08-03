package fetcher

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/Masterminds/squirrel"

	"github.com/jsocol/dataloader/examples/grpc-resource-server/proto"
)

type bookRecord struct {
	ID    string `sql:"id"`
	Title string `sql:"title"`
}

type fetcher struct {
	sq squirrel.StatementBuilderType
	db *sql.DB
}

func New(db *sql.DB) *fetcher {
	stmtCache := squirrel.NewStmtCache(db)
	return &fetcher{
		sq: squirrel.StatementBuilder.RunWith(stmtCache),
		db: db,
	}
}

func (f *fetcher) Fetch(ids []string) (map[string]*proto.Book, error) {
	ctx := context.TODO()

	query := f.sq.Select("id", "title").From("books")
	if len(ids) > 0 {
		query = query.Where(squirrel.Eq{"id": ids})
	}

	str, args, _ := query.ToSql()
	slog.DebugContext(ctx, "SELECT", "ids", ids, "query", str, "args", args)

	rows, err := query.QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ret := make(map[string]*proto.Book, len(ids))

	for rows.Next() {
		var book bookRecord
		if err = rows.Scan(&book.ID, &book.Title); err != nil {
			return nil, err
		}
		ret[book.ID] = &proto.Book{
			Id:    book.ID,
			Title: book.Title,
		}
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return ret, nil
}
