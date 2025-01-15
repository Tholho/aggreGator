This Project was done with the guidance of boot.dev.

It is an RSSFlux aggre-gator. Hence the name. The joke is not mine so spare me please.

The concept is to automatically fetch RSS feeds from a list of followed feeds. It allows multipler users to have their own selection of different feeds, but allows for each user to consult the existing feeds in the db and to follow them if they wish to.

To run this software, unless I provide a built package (which really is not something anyone would require as far as I reckon), you'll need Golang. Additionnally, you need to have postgres installed. 

**tl;dr : Need PostGres + Golang**

Every feature is pretty basic, but those basics are covered as follows :

* "register" : ie. 'aggreGator register "username"' adds the user to the db and unlocks other commands. Automatically log in the newly registered user
* "login" : ie. 'aggreGator login "username"' | Provided user is registered, will run the following commands for the current user
* "reset" : admin feature, clears the users data
* "users" : lists all users
* "agg" : ie. 'aggreGator agg *time duration*' (such as 1s if you want to DDOS a server, or 5 min if you're reasonable but addicted to news), will refresh the posts db (and make it grow, no cleanup feature yet)    
* "addfeed" : ie. 'aggreGator addfeed *url*' | Will add a new RSSfeed AND automatically follow it for the current user
* "feeds" : lists existing feeds
* "follow" : ie. 'aggreGator follow *url*' | Follow an already stored feed
* "unfollow" : reverts the previous command, url required. (unfollow *url*)
* "following" :  lists existing follows for the current user
* "browse" : optional limit parameter | ie 'aggreGator browse' OR 'browse *limit number*' | will display a default 2 (or provided *limit number*) most recent posts from all followed feeds for the current user

If you don't know how to install Go software I'm supposed to tell you but nothing beats the original so here you go :

https://pkg.go.dev/cmd/go#hdr-Compile_and_install_packages_and_dependencies

**tl;dr : Assuming you installed postgres and golang... Download files, run "go install" from folder, then run aggreGator**

**Last but not least**

you need a .gatorconfig.json file at your ~HOME folder, its content is as follows :

{
  "db_url": "postgres://**user**:*password*@localhost:5432/gator?sslmode=disable",
  "current_user_name": "bob"
}

You should replace **user** and *password* fields with the ones you chose during postgres install. I think postgres is the default user.
