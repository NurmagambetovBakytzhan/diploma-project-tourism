import pandas as pd
from sklearn.preprocessing import LabelEncoder, OneHotEncoder, normalize
from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.metrics.pairwise import cosine_similarity
from scipy.sparse import hstack

def prepare_recommendation_system(tour_data):
    le = LabelEncoder()
    tour_data['category_encoded'] = le.fit_transform(tour_data['category'])

    tfidf = TfidfVectorizer()
    tour_tfidf_matrix = tfidf.fit_transform(tour_data['description'])
    tfidf_matrix_normalized = normalize(tour_tfidf_matrix)

    onehot = OneHotEncoder()
    category_matrix = onehot.fit_transform(tour_data[['category_encoded']])
    category_matrix_normalized = normalize(category_matrix)

    tfidf_weight = 0.7
    category_weight = 0.3

    tfidf_matrix_weighted = tfidf_matrix_normalized * tfidf_weight
    category_matrix_weighted = category_matrix_normalized * category_weight

    combined_features = hstack([tfidf_matrix_weighted, category_matrix_weighted])
    cosine_sim_combined = cosine_similarity(combined_features, combined_features)

    return tour_data, cosine_sim_combined

def recommend_for_user(user_visited_ids, tour_data, cosine_sim_matrix, top_n=5):
    if not user_visited_ids:
        sampled = tour_data.groupby('category').apply(lambda x: x.sample(1)).reset_index(drop=True)
        return sampled[['id', 'name', 'description', 'category']].head(top_n).to_dict(orient='records')
    visited_indices = tour_data[tour_data['id'].isin(user_visited_ids)].index
    sim_scores = cosine_sim_matrix[visited_indices].mean(axis=0)

    sim_indices = sim_scores.argsort()[::-1]
    recommended_indices = [i for i in sim_indices if tour_data.iloc[i]['id'] not in user_visited_ids]

    return tour_data.iloc[recommended_indices].head(top_n)[['id', 'name', 'description', 'category']].to_dict(orient='records')
