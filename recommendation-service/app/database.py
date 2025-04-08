from sqlalchemy import create_engine, text
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker, Session

from .config import settings
import pandas as pd

SQLALCHEMY_DATABASE_URL = (f'postgresql://{settings.database_username}:{settings.database_password}@'
                           f'{settings.database_hostname}:{settings.database_port}/{settings.database_name}')



engine = create_engine(SQLALCHEMY_DATABASE_URL)

SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)

Base = declarative_base()


def get_db():
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()

def fetch_tour_data(db:Session):
    query = """
        SELECT tourism.tours.ID, tourism.tours.description, tourism.tours.name, tourism.categories.name AS category
        FROM tourism.tours
        INNER JOIN tourism.tour_categories ON tours.id = tour_categories.tour_id
        INNER JOIN tourism.categories ON tour_categories.category_id = tourism.categories.id
    """
    result = db.execute(text(query))
    df = pd.DataFrame(result.fetchall(), columns=result.keys())
    return df

def fetch_user_activities(db: Session):
    query = "SELECT user_id, tour_id FROM tourism.user_activities"
    result = db.execute(text(query))
    df = pd.DataFrame(result.fetchall(), columns=result.keys())
    return df
# while True:
#     try:
#         conn = psycopg2.connect(host='localhost', database='fastapi', user='postgres', password='123',
#                                 cursor_factory=RealDictCursor)
#         cursor = conn.cursor()
#         print("Database conn successfull")
#         break
#     except Exception as error:
#         print("Connection failed")
#         print(error)
#         time.sleep(2)
