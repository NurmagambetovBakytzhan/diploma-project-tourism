from uuid import UUID

from fastapi import FastAPI, HTTPException, Depends

from fastapi.middleware.cors import CORSMiddleware
from sqlalchemy.orm import Session

from .config import settings
from .database import fetch_tour_data, fetch_user_activities, get_db
from .recommender import prepare_recommendation_system, recommend_for_user

print(settings.database_username)

# models.Base.metadata.create_all(bind=engine)

app = FastAPI()

origins = ["*"]

app.add_middleware(
    CORSMiddleware,
    allow_origins=origins,
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


@app.get("/recommendations/health")
async def root():
    return {"message": "Recomendation Service!"}

@app.get("/recommendations/{user_id}")
def get_user_recommendations(user_id: UUID, db: Session = Depends(get_db)):
    try:
        tour_data = fetch_tour_data(db)
        user_activity = fetch_user_activities(db)
        print(tour_data)
        print("!~~~!")
        print(user_activity)
        tour_data, cosine_sim_matrix = prepare_recommendation_system(tour_data)

        user_visited_ids = user_activity[user_activity['user_id'] == user_id]['tour_id'].tolist()

        recommendations = recommend_for_user(user_visited_ids, tour_data, cosine_sim_matrix)

        return {"user_id": user_id, "recommendations": recommendations}

    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Error: {str(e)}")
