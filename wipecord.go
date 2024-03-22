package main

import (
    "fmt"
    "flag"
    "os"
    "encoding/json"
    "net/http"
    "bufio"
    "strings"
    "time"
    "io"
)

// Constants
var DISCORD_API string = "https://discord.com/api/v9"
type authed_client struct {
    client http.Client
    token* string
}
type message struct {
    Channel_Id string `json:"channel_id"`
    Message_Id string `json:"id"`
    Author struct { Id string `json:"id"` } `json:"author"`
    Content string `json:"content"`
}

func main(){
    // Parse command line arguments
    token := flag.String("t", os.Getenv("DISCORD_TOKEN"), "Discord Token")
    guild_id := flag.String("g", "", "Guild ID")
    author := flag.String("a", "@me", "Author ID")
    yes := flag.Bool("y", false, "Skip confirmations")
    delete_rate := flag.Int("r", 1500, "Delay between message deletions in milliseconds")
    fetch_rate := flag.Int("f", 30_000, "Delay between message fetches in milliseconds")
    flag.Parse()

    // Abort w/ usage if no arguments
    if flag.NFlag() == 0 {
        abort("No arguments provided")
    }

    var ticket authed_client = authed_client{http.Client{}, token}
    guild_name := get_guild(&ticket, guild_id)
    author_id, author_name := get_author(&ticket, author)
    total_messages := get_total_messages(&ticket, guild_id, &author_id)

    // Confirm wipe
    fmt.Printf("Wipe %d messages sent by @%s in \"%s\"? [y/N] ", total_messages, author_name, guild_name)
    if !*yes {
        reader := bufio.NewReader(os.Stdin)
        in, err := reader.ReadString('\n')
		if err != nil { abort(err.Error()) }

		in = strings.ToLower(strings.TrimSpace(in))

		if in != "y" && in != "yes" {
		    os.Exit(0)
		}
    }else{
        fmt.Println("y")
    }

    // Wipe messages
    fmt.Println("Wiping messages...")

    for {
        count, messages := get_message_bundle(&ticket, guild_id, &author_id)
        if count == 0 { break }
        for _, message := range messages {
            if delete_message(&ticket, &message, *delete_rate){
                fmt.Println(" Deleted message:", message.Content)
            }
        }
        fmt.Println("Deleted", fmt.Sprintf("%d/%d", total_messages - count + 20, total_messages), "messages so far")
        fmt.Println("\nWaiting", *fetch_rate, "ms before fetching more messages...")
        time.Sleep(time.Duration(*fetch_rate) * time.Millisecond)
    }
}

func delete_message(ticket* authed_client, message* message, rate int) bool {
    req := discord(ticket, "DELETE", "/channels/" + message.Channel_Id + "/messages/" + message.Message_Id)
    time.Sleep(time.Duration(rate) * time.Millisecond)
    if req.StatusCode == 204 { return true }

    // Message deletion failed
    body, err := io.ReadAll(req.Body)
    if err != nil { abort(err.Error()) }
    fmt.Println("Failed to delete message", message.Message_Id, ":", string(body))
    return false
}

func get_message_bundle(ticket* authed_client, guild_id* string, author_id* string) (int, []message) {
    var bundle struct { 
        Count int `json:"total_results"`
        Messages [][]message `json:"messages"`
    }
    bundle_req := discord(ticket, "GET", "/guilds/" + *guild_id + "/messages/search?author_id=" + *author_id + "&sort_order=asc&include_nsfw=true")
    bundle_err := json.NewDecoder(bundle_req.Body).Decode(&bundle)
    if bundle_err != nil { abort(bundle_err.Error()) }

    // Poor man's list comprehension
    var message_ids []message
    for _, m := range bundle.Messages {
        if m[0].Author.Id == *author_id {
            message_ids = append(message_ids, m[0])
        }
    }

    fmt.Println("Fetched", len(message_ids), "messages")
    return bundle.Count, message_ids
}

func get_total_messages(ticket* authed_client, guild_id* string, author_id* string) int {
    var count struct { Total int `json:"total_results"` }
    count_req := discord(ticket, "GET", "/guilds/" + *guild_id + "/messages/search?author_id=" + *author_id + "&include_nsfw=true")
    count_err := json.NewDecoder(count_req.Body).Decode(&count)
    if count_err != nil { abort(count_err.Error()) }
    return count.Total
}

func get_guild(ticket* authed_client, guild_id* string) string {
    var guild struct { Name string `json:"name"` }
    guild_req := discord(ticket, "GET", "/guilds/" + *guild_id)
    guild_err := json.NewDecoder(guild_req.Body).Decode(&guild)
    if guild_err != nil { abort(guild_err.Error()) }
    return guild.Name
}

func get_author(ticket* authed_client, author_id* string) (string, string) {
    var author struct { Id string `json:"id"`; Name string `json:"username"` }
    author_req := discord(ticket, "GET", "/users/" + *author_id)
    author_err := json.NewDecoder(author_req.Body).Decode(&author)
    if author_err != nil { abort(author_err.Error()) }
    return author.Id, author.Name
}

func discord(ticket* authed_client, req_type string, url string) *http.Response {
    req, err := http.NewRequest(req_type, DISCORD_API + url, nil)
    req.Header.Add("Content-Type", "application/json")
    req.Header.Add("Authorization", *ticket.token)
    if err != nil {
        abort(err.Error())
    }

    resp, err := ticket.client.Do(req)
    if err != nil {
        abort(err.Error())
    }

    return resp
}

func abort(err ...string){
    if len(err) > 0 {
        fmt.Println("Error: " + err[0])
    }
    flag.Usage()
    os.Exit(1)
}

