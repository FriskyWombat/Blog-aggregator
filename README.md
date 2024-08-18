# Blog-aggregator

## Overview
Blog Aggregator is a web API that allows users to add blogs to a PostgreSQL database, follow any blogs, and retrieve posts from their followed blogs. This was a guided project through Boot.Dev.

## Features
- **Add Blogs**: Users can add new blogs to the database.
- **Follow Blogs**: Users can follow any blogs they are interested in.
- **Retrieve Posts**: Users can retrieve and read posts from the blogs they follow.

## Endpoints

### Health Check
- **GET /healthz**: Check if the service is running.

### Error Testing
- **GET /err**: Trigger an error (for testing purposes).

### User Management
- **POST /users**: Create a new user. Response contains an API key necessary for authenticated requests.
- **GET /users**: Retrieve user information (requires authentication).

### Blog Feeds
- **POST /feeds**: Add a new blog feed (requires authentication).
- **GET /feeds**: Retrieve all blog feeds.

### Follow Management
- **POST /feed_follows**: Follow a blog feed (requires authentication).
- **GET /feed_follows**: Retrieve followed blog feeds (requires authentication).
- **DELETE /feed_follows/{id}**: Unfollow a blog feed by ID (requires authentication).

### Posts
- **GET /posts**: Retrieve posts from followed blogs (requires authentication).