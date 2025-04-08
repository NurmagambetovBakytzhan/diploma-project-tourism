from uuid import UUID

from fastapi import FastAPI, HTTPException, Depends

from fastapi.middleware.cors import CORSMiddleware
from sqlalchemy.orm import Session

from .config import settings
from .database import fetch_tour_data, fetch_user_activities, get_db
from .recommender import prepare_recommendation_system, recommend_for_user

# models.Base.metadata.create_all(bind=engine)

app = FastAPI(
    title="Recommendation Service API",
    description="API for generating personalized tour recommendations",
    version="1.0.0",
    openapi_url="/v1/recommendations/openapi.json",
    docs_url="/v1/recommendations/docs",
    redoc_url="/v1/recommendations/redoc"
)

origins = ["*"]

app.add_middleware(
    CORSMiddleware,
    allow_origins=origins,
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


@app.get("/v1/recommendations/health")
async def root():
    return {"message": "Recomendation Service!"}


@app.get("/v1/recommendations/{user_id}")
def get_user_recommendations(user_id: UUID, db: Session = Depends(get_db)):
    try:
        tour_data = fetch_tour_data(db)
        user_activity = fetch_user_activities(db)
        tour_data, cosine_sim_matrix = prepare_recommendation_system(tour_data)

        user_visited_ids = user_activity[user_activity['user_id'] == user_id]['tour_id'].tolist()

        recommendations = recommend_for_user(user_visited_ids, tour_data, cosine_sim_matrix)

        return {"user_id": user_id, "recommendations": recommendations}

    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Error: {str(e)}")
