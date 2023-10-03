import React, { useState, useEffect } from "react";
import { db } from "../../firebase";
import PetitionModal from "../../components/PetitionModal";
import { Link } from "react-router-dom";

function AllBlogs() {
  const [blogs, setBlogs] = useState([]);
  useEffect(() => {
    db.collection("Blogs")
      .get()
      .then((blogsSnapshot) => {
        const blogsData = blogsSnapshot.docs.map((doc) => ({
          id: doc.id,
          ...doc.data(),
        }));
        setBlogs(blogsData);
      })
      .catch((error) => {
        console.error("Error fetching blogs data:", error);
      });
  }, []);

  return (
    <div>
      <section className="blog-section">
        <h2 className="section-title">All Blog Posts</h2>
        <div className="blog-grid">
          {blogs &&
            blogs.map((post, index) => (
              <Link className="blog-post" key={index} to={`/blog/${post.id}`}>
                <h3 className="post-title">
                  {post.title.length > 50
                    ? post.title.substring(0, 50) + "..."
                    : post.title}
                </h3>
                <p className="post-content">
                  {post.content.length > 250
                    ? post.content.substring(0, 250) + "..."
                    : post.content}
                </p>
                <p className="post-author">By {post?.ownerName}</p>
                <p className="post-date">
                  {post.timestamp.toDate().toDateString()}
                </p>
              </Link>
            ))}
        </div>
      </section>
    </div>
  );
}

export default AllBlogs;
