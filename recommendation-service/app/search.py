from fastapi import Depends
from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.metrics.pairwise import cosine_similarity
import numpy as np
from pgvector.psycopg import register_vector
import os
import psycopg
from sqlalchemy import text
from sqlalchemy.orm import Session
from .database import fetch_tour_data, fetch_user_activities, get_db

model_name = "multi-qa-MiniLM-L6-cos-v1"


def prepare_semantic_search(model, tour_data):
    conn = psycopg.connect(
        host='tourism-db',
        dbname='tourism-db',
        user='postgres',
        password='postgres',
        port='5432'
    )
    conn.execute("SET search_path TO tourism, public")

    register_vector(conn)
    cur = conn.cursor()
    for _, row in tour_data.iterrows():
        description = row['description'] or ""
        category = row['category'] or ""
        name = row['name'] or ""

        emb = model.encode(f"{description} {category} {name}")
        embedding_str = "[" + ",".join(map(str, emb)) + "]"

        cur.execute(
            """
            INSERT INTO tourism.tour_embeddings (tour_id, embedding)
            VALUES (%s, %s)
            """,
            (row['id'], embedding_str)
        )
    conn.commit()


def semantic_search(query, model, size, offset):
    conn = psycopg.connect(
        host='tourism-db',
        dbname='tourism-db',
        user='postgres',
        password='postgres',
        port='5432'
    )

    conn.execute("SET search_path TO tourism, public")

    register_vector(conn)
    emb = model.encode(query)

    rows = conn.execute(
        """
        SELECT 
            te.embedding <=> %s AS distance,
            t.id,
            t.name,
            t.description,
            array_agg(DISTINCT i.image_url) AS image_urls,
            array_agg(DISTINCT c.name) AS categories
        FROM tourism.tour_embeddings te
        INNER JOIN tourism.tours t ON te.tour_id = t.id
        LEFT JOIN tourism.images i ON t.id = i.tour_id
        LEFT JOIN tourism.tour_categories tc ON t.id = tc.tour_id
        LEFT JOIN tourism.categories c ON tc.category_id = c.id
        GROUP BY t.id, te.embedding
        ORDER BY te.embedding <=> %s
        LIMIT %s OFFSET %s
        """,
        (emb, emb, size, offset)
    ).fetchall()

    results = [
        {
            "distance": row[0],
            "id": row[1],
            "name": row[2],
            "description": row[3],
            "image_urls": row[4],
            "categories": row[5]
        }
        for row in rows
    ]
    return results