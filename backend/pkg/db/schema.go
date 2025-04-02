package db

var (
	UsersTable = `CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,  
		username TEXT UNIQUE NOT NULL,       
		email TEXT UNIQUE NOT NULL,          
		password_hash TEXT NOT NULL,          
		first_name TEXT NOT NULL,            
		last_name TEXT NOT NULL,              
		role TEXT CHECK(role IN ('admin', 'moderator', 'user')) DEFAULT 'user',
		gender TEXT CHECK(gender IN ('Homme', 'Femme', 'autre')), 
		date_of_birth DATE NOT NULL,          
		avatar TEXT,                          
		bio TEXT,                             
		phone_number TEXT UNIQUE,             
		address TEXT,                         
		is_private BOOLEAN DEFAULT FALSE,     
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, 
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP  
	);`

	FollowersTable = `CREATE TABLE IF NOT EXISTS followers (
		id TEXT PRIMARY KEY,
		follower_id TEXT NOT NULL,
		followed_id TEXT NOT NULL,
		status TEXT CHECK(status IN ('pending', 'accepted')) DEFAULT 'pending',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (follower_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (followed_id) REFERENCES users(id) ON DELETE CASCADE
	);`

	FollowRequestsTable = `CREATE TABLE IF NOT EXISTS follow_requests (
		id TEXT PRIMARY KEY,
		sender_id TEXT NOT NULL,
		receiver_id TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (sender_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (receiver_id) REFERENCES users(id) ON DELETE CASCADE
	);`

	GroupsTable = `CREATE TABLE IF NOT EXISTS groups (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		description TEXT,
		creator_id TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (creator_id) REFERENCES users(id) ON DELETE CASCADE
	);`

	GroupMembersTable = `CREATE TABLE IF NOT EXISTS group_members (
		id TEXT PRIMARY KEY,
		group_id TEXT NOT NULL,
		user_id TEXT NOT NULL,
		status TEXT CHECK(status IN ('pending', 'accepted')) DEFAULT 'pending',
		role TEXT CHECK(role IN ('creator', 'member')) DEFAULT 'member',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);`

	GroupPosts = `CREATE TABLE IF NOT EXISTS group_posts (
		id TEXT PRIMARY KEY,
		group_id TEXT NOT NULL,
		user_id TEXT NOT NULL,
		title TEXT NOT NULL,
		content TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);`

	GroupPostsComments = `CREATE TABLE IF NOT EXISTS group_posts_comments (
		id TEXT PRIMARY KEY,
		post_id TEXT NOT NULL,
		content TEXT NOT NULL,
		user_id TEXT NOT NULL,
		username TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (post_id) REFERENCES group_posts(id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);`

	GroupEvents = `CREATE TABLE IF NOT EXISTS group_events (
		id TEXT PRIMARY KEY,
		group_id TEXT NOT NULL,
		user_id TEXT NOT NULL,
		title TEXT NOT NULL,
		description TEXT NOT NULL,
		event_date DATETIME NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);`

	GroupsMessage = `CREATE TABLE group_messages (
		id UUID PRIMARY KEY,
		group_id UUID REFERENCES groups(id),
		sender_id UUID REFERENCES users(id),
		content TEXT,
		emoji TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	EventResponses = `CREATE TABLE IF NOT EXISTS event_responses (
		id TEXT PRIMARY KEY,
		event_id TEXT NOT NULL,
		user_id TEXT NOT NULL,
		response TEXT CHECK(response IN ('Going', 'Not going')) NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (event_id) REFERENCES group_events(id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);`

	NotificationsTable = `CREATE TABLE IF NOT EXISTS notifications (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,          
		content TEXT NOT NULL,             
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		read BOOLEAN DEFAULT FALSE,        
		type TEXT CHECK(type IN ('follow_request', 'follow_accept', 'new_post', 'new_comment', 'message')) DEFAULT 'new_post',
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);`

	MessagesTable = `CREATE TABLE IF NOT EXISTS messages (
		id TEXT PRIMARY KEY,
		sender_id TEXT NOT NULL,
		recipient_id TEXT NOT NULL,
		content TEXT NOT NULL,
		emoji TEXT,
		image_path TEXT,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (sender_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (recipient_id) REFERENCES users(id) ON DELETE CASCADE
	);`

	PostsTable = `CREATE TABLE IF NOT EXISTS posts (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		content TEXT NOT NULL,
		user_id TEXT NOT NULL,
		visibility TEXT CHECK(visibility IN ('public', 'private', 'almost_private')) DEFAULT 'public',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		image_path TEXT,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);`

	PostsTableAllowed = `CREATE TABLE IF NOT EXISTS post_allowed_users (
		post_id TEXT NOT NULL,
		user_id TEXT NOT NULL,
		FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);`

	CommentsTable = `CREATE TABLE IF NOT EXISTS comments (
		id TEXT PRIMARY KEY,
		post_id TEXT NOT NULL,
		content TEXT NOT NULL,
		user_id TEXT NOT NULL,
		username TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);`

	CommentInteractionsTable = `CREATE TABLE IF NOT EXISTS comment_interactions (
		id TEXT PRIMARY KEY,
		comment_id TEXT NOT NULL,
		user_id TEXT NOT NULL,
		interaction_type TEXT CHECK(interaction_type IN ('like', 'unlike')) NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (comment_id) REFERENCES comments(id) ON DELETE CASCADE
	);`

	PostInteractionsTable = `CREATE TABLE IF NOT EXISTS post_interactions (
		id TEXT PRIMARY KEY,
		post_id TEXT NOT NULL,
		user_id TEXT NOT NULL,
		interaction_type TEXT CHECK(interaction_type IN ('like', 'unlike')) NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE
	);`
)
