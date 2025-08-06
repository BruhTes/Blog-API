package repository

import (
	"Blog-API/internal/domain"
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"Blog-API/internal/infrastructure/database"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type BlogRepo struct {
	db         *database.MongoDB
	collection *mongo.Collection
}

// CORRECTED: The constructor now returns the interface type and takes the standard *mongo.Database.
func NewBlogRepository(db *database.MongoDB) *BlogRepo {
	// CORRECTED: Collection names are conventionally lowercase.
	collection := db.GetCollection("blogs")
	return &BlogRepo{db: db, collection: collection}
}

func (br *BlogRepo) Create(blog *domain.Blog) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := br.collection.InsertOne(ctx, blog)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("blog with this ID already exists: %w", err)
		}
		if mongo.IsTimeout(err) {
			return fmt.Errorf("database operation timed out: %w", err)
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("database context deadline exceeded: %w", err)
		}
		return fmt.Errorf("failed to create blog: %w", err)
	}
	return nil
}

func (br *BlogRepo) GetByID(id primitive.ObjectID) (*domain.Blog, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var blog domain.Blog
	filter := bson.M{"_id": id}

	err := br.collection.FindOne(ctx, filter).Decode(&blog)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("blog not found")
		}
		return nil, fmt.Errorf("database error in GetByID: %w", err)
	}
	return &blog, nil
}

func (br *BlogRepo) GetAll(page, limit int, sort string) ([]*domain.Blog, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var blogs []*domain.Blog
	// CORRECTED: An empty BSON document matches all.
	filter := bson.D{}
	opts := options.Find()
	opts.SetLimit(int64(limit))
	opts.SetSkip(int64(page-1) * int64(limit))

	// A simple sort implementation
	if sort == "popular" {
		opts.SetSort(bson.D{{Key: "like_count", Value: -1}})
	} else {
		opts.SetSort(bson.D{{Key: "created_at", Value: -1}})
	}

	curr, err := br.collection.Find(ctx, filter, opts)
	if err != nil {
		// No need for log.Fatal, just return the error
		return nil, 0, fmt.Errorf("failed to find blogs: %w", err)
	}
	defer curr.Close(ctx)

	if err := curr.All(ctx, &blogs); err != nil {
		return nil, 0, err
	}

	total, err := br.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return blogs, total, nil
}

// CORRECTED: This logic is now simple and correct.
func (br *BlogRepo) Update(blog *domain.Blog) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": blog.ID}
	update := bson.M{"$set": blog}

	result, err := br.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update blog: %w", err)
	}

	if result.MatchedCount == 0 {
		return errors.New("blog not found for update")
	}

	return nil
}

func (br *BlogRepo) Delete(id primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := br.collection.DeleteOne(ctx, bson.D{{Key: "_id", Value: id}})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return errors.New("blog not found for delete")
	}
	return nil
}

func (br *BlogRepo) SearchByTitle(title string, page, limit int) ([]*domain.Blog, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var blogs []*domain.Blog
	// Note: For best results, a text index should be created on this field in MongoDB.
	filter := bson.M{"title": bson.M{"$regex": title, "$options": "i"}} // Case-insensitive substring search

	// Re-using a helper for paginated queries would be ideal, but for now this is fine.
	opts := options.Find()
	opts.SetLimit(int64(limit))
	opts.SetSkip(int64(page-1) * int64(limit))

	curr, err := br.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer curr.Close(ctx)

	if err := curr.All(ctx, &blogs); err != nil {
		return nil, 0, err
	}

	total, err := br.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return blogs, total, nil
}

func (br *BlogRepo) SearchByAuthor(author string, page, limit int) ([]*domain.Blog, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var blogs []*domain.Blog
	// CORRECTED: The field name must match the schema exactly.
	filter := bson.M{"author_username": author}

	opts := options.Find()
	opts.SetLimit(int64(limit))
	opts.SetSkip(int64(page-1) * int64(limit))

	curr, err := br.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer curr.Close(ctx)

	if err := curr.All(ctx, &blogs); err != nil {
		return nil, 0, err
	}

	total, err := br.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return blogs, total, nil
}

// --- ADDED STUB IMPLEMENTATIONS FOR ALL MISSING METHODS ---
// These are required for the code to compile. They return empty data.

func (Br *BlogRepo) FilterByTags(tags []string, page, limit int) ([]*domain.Blog, int64, error){
	ctx , cancle := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancle()
	sort := bson.D{{"title",1}}
	opt := options.Find().SetSort(sort)
	opt.SetLimit(int64(limit))	
	opt.SetSkip( int64(page - 1) * int64(limit))

	filter := bson.D{{"tags",bson.D{{"$in",tags}}}}
	var blogs []*domain.Blog

	curr, err := Br.collection.Find(ctx,filter,opt)
	if err != nil {
		log.Println("Find error:", err)
		return nil, 0, err
	}

	if err := curr.All(ctx,&blogs); err != nil{
		log.Println("Find error:", err)
		return nil, 0, err
	}

	total, err := Br.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return blogs,total,nil


}
func  (Br *BlogRepo)FilterByDate(startDate, endDate time.Time, page, limit int) ([]*domain.Blog, int64, error){

	ctx , cancle := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancle()
	
	sort := bson.D{{"title",1}}
	opt := options.Find().SetSort(sort)
	opt.SetLimit(int64(limit))	
	opt.SetSkip( int64(page - 1) * int64(limit))

	filter := bson.D{{"created_at", bson.D{{"$gte", startDate},{"$lte", endDate},}},}

	var blogs []*domain.Blog

	curr, err := Br.collection.Find(ctx,filter,opt)
	if err != nil {
		log.Println("Find error:", err)
		return nil, 0, err
	}

	if err := curr.All(ctx,&blogs); err != nil{
		log.Println("Find error:", err)
		return nil, 0, err
	}

	total, err := Br.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return blogs,total,nil
}

func (Br *BlogRepo) GetPopular(limit int) ([]*domain.Blog, error){

	ctx , cancle := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancle()
	
	sort := bson.D{{"view_count",1}}
	opt := options.Find().SetSort(sort)
	opt.SetLimit(int64(limit))		

	var blogs []*domain.Blog

	curr, err := Br.collection.Find(ctx,bson.D{{}},opt)
	if err != nil {
		log.Println("Find error:", err)
		return nil,  err
	}

	if err := curr.All(ctx,&blogs); err != nil{
		log.Println("Find error:", err)
		return nil,  err
	}

	
	return blogs,nil



}



func (Br *BlogRepo) IncrementViewCount(id primitive.ObjectID) error{
	ctx, cancle := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancle()

	_, err := Br.collection.UpdateOne(ctx,bson.D{{"_id",id}},bson.D{{"$inc",bson.D{{"view_count", 1}}}})

	if err != nil{
		return  err
	}
	return  nil

}
func (Br *BlogRepo) AddComment(blogID primitive.ObjectID, comment *domain.Comment) error{
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Set creation and update time
	comment.ID = primitive.NewObjectID()
	comment.CreatedAt = time.Now()
	comment.UpdatedAt = time.Now()

	update := bson.D{
		{"$push", bson.D{
			{"comments", comment},
		}},
		{"$inc", bson.D{
			{"comment_count", 1}, // Optional: increment comment count
		}},
	}

	_, err := Br.collection.UpdateOne(ctx, bson.M{"_id": blogID}, update)
	if err != nil {
		return err
	}

	return nil
}
func (Br *BlogRepo)DeleteComment(blogID, commentID primitive.ObjectID) error{
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.D{
		{"$pull",bson.D{{"comments",bson.D{{"_id",commentID}}}}},
		{"$inc",bson.D{{"comment_count",-1}}},
	}

	filter := bson.D{{"_id",blogID}}

	_,err := Br.collection.UpdateOne(ctx,filter,update)

	if err != nil{
		return  err
	}

	return nil

}

func (Br *BlogRepo) UpdateComment(blogID, commentID primitive.ObjectID, content string) error{
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()


	filter := bson.D{
		{"_id", blogID},
		{"comments._id", commentID},
	}

	
	update := bson.D{
		{"$set", bson.D{
			{"comments.$.content", content},
			{"comments.$.updated_at", time.Now()},
		}},
	}

	result, err := Br.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.ModifiedCount == 0 {
		return fmt.Errorf("no comment updated — check comment ID")
	}

	return nil

	
}
func (Br *BlogRepo)AddLike(blogID primitive.ObjectID, userID string) error{
	ctx, cancle := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancle()

	_, err := Br.collection.UpdateOne(ctx,bson.D{{"_id",blogID}},bson.D{{"$inc",bson.D{{"like_count", 1}}},{"$push",bson.D{{"likes",userID}}}})

	if err != nil{
		return  err
	}
	return  nil
}
func (Br *BlogRepo)RemoveLike(blogID primitive.ObjectID, userID string) error{
	ctx, cancle := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancle()

	_, err := Br.collection.UpdateOne(ctx,bson.D{{"_id",blogID}},bson.D{{"$inc",bson.D{{"like_count", -1}}},{"$pull",bson.D{{"likes",userID}}}})

	if err != nil{
		return  err
	}
	return  nil

}
func (Br *BlogRepo) AddDislike(blogID primitive.ObjectID, userID string) error{
	ctx, cancle := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancle()

	_, err := Br.collection.UpdateOne(ctx,bson.D{{"_id",blogID}},bson.D{{"$push",bson.D{{"dislikes",userID}}}})

	if err != nil{
		return  err
	}
	return  nil

}
func (Br * BlogRepo)RemoveDislike(blogID primitive.ObjectID, userID string) error{
	ctx, cancle := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancle()

	_, err := Br.collection.UpdateOne(ctx,bson.D{{"_id",blogID}},bson.D{{"$pull",bson.D{{"dislikes",userID}}}})

	if err != nil{
		return  err
	}
	return  nil

	}

// func (br *BlogRepo) GetTagIDByName(name string) (primitive.ObjectID, error) {
//  return primitive.NilObjectID, nil
// }