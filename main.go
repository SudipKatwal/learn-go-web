package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var db *mongo.Database
var ctx = context.TODO()

type (
	Todo struct {
		ID        primitive.ObjectID `bson:"_id"`
		Title     string             `bson:"title"`
		CreatedAt time.Time          `bson:"createdAt"`
	}

	pageData struct {
		Title string
		Todos []*Todo
	}
)

func init() {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017/")
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	db = client.Database("db_todo")

}

func main() {
	http.HandleFunc("/", todoPageHandler)
	http.HandleFunc("/create-todo", createdTodoPageHandler)
	http.HandleFunc("/delete", deleteTodoPageHandler)
	http.HandleFunc("/update", updateTodoPageHandler)
	http.ListenAndServe(":5000", nil)
}

func updateTodoPageHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	id := r.FormValue("id")
	title := r.FormValue("title")

	collection := db.Collection("todo")

	pId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Fatal(err)
	}
	res, err := collection.UpdateOne(ctx, bson.M{"_id": bson.M{"$eq": pId}}, bson.M{"$set": bson.M{"title": title}})
	if err != nil {
		log.Fatal(err)
		return
	}
	fmt.Println(res.UpsertedID)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func deleteTodoPageHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	id := r.FormValue("id")

	collection := db.Collection("todo")

	pId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Fatal(err)
	}
	res, err := collection.DeleteOne(ctx, bson.M{"_id": pId})
	if err != nil {
		log.Fatal(err)
		return
	}
	fmt.Println(res.DeletedCount)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func todoPageHandler(w http.ResponseWriter, r *http.Request) {
	var tasks []*Todo
	collection := db.Collection("todo")

	cur, err := collection.Find(ctx, bson.D{})
	if err != nil {
		log.Fatal(err)
	}

	for cur.Next(ctx) {
		var result Todo
		err := cur.Decode(&result)
		if err != nil {
			log.Fatal(err)
		}
		// do something with result....
		tasks = append(tasks, &result)
	}
	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}
	p := pageData{
		Title: "Todo Web App",
		Todos: tasks,
	}
	t, _ := template.ParseFiles("todo.html")
	t.Execute(w, p)
}

func createdTodoPageHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	title := r.FormValue("title")

	collection := db.Collection("todo")

	todo := &Todo{
		ID:        primitive.NewObjectID(),
		Title:     title,
		CreatedAt: time.Now(),
	}

	res, err := collection.InsertOne(ctx, todo)
	if err != nil {
		log.Fatal(err)
		return
	}
	fmt.Println(res.InsertedID)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
