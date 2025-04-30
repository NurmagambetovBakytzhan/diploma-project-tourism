from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.metrics.pairwise import cosine_similarity
import numpy as np


def prepare_semantic_search(tour_data):
    # Combine relevant fields more thoughtfully
    tour_data['search_text'] = (
            tour_data['name'] + " " +
            tour_data['description'] + " " +
            tour_data['category']
    )

    # Create TF-IDF vectorizer with better parameters
    vectorizer = TfidfVectorizer(
        ngram_range=(1, 2),  # Include unigrams and bigrams
        min_df=2,  # Ignore terms that appear in only 1 document
        max_df=0.85,  # Ignore terms that appear in >85% of documents
        sublinear_tf=True  # Use sublinear TF scaling
    )

    tfidf_matrix = vectorizer.fit_transform(tour_data['search_text'])
    return tour_data, tfidf_matrix, vectorizer


def semantic_search(query, tour_data, tfidf_matrix, vectorizer, top_n=5):
    # Preprocess query similarly to how documents were processed
    processed_query = query.lower().strip()

    # Vectorize the query using the same vectorizer
    query_vec = vectorizer.transform([processed_query])

    # Compute similarities
    cos_sim = cosine_similarity(query_vec, tfidf_matrix).flatten()

    # Get top results with proper sorting
    top_indices = np.argsort(cos_sim)[-top_n:][::-1]  # Get top N sorted descending

    results = []
    for idx in top_indices:
        # Skip results with zero similarity
        if cos_sim[idx] <= 0:
            continue

        results.append({
            'id': tour_data.iloc[idx]['id'],
            'name': tour_data.iloc[idx]['name'],
            'description': tour_data.iloc[idx]['description'],
            'category': tour_data.iloc[idx]['category'],
            'score': float(cos_sim[idx])  # Convert numpy float to Python float
        })

    # If no results with positive similarity, return empty
    return results if results else []