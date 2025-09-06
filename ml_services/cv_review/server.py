from fastapi import FastAPI

from reviewer_model import GPTReviewer


app = FastAPI(title='CV analysis')

@app.get("/get_review")
async def review_cv(vacancy: str, cv: str, skillvals: str):
    model = GPTReviewer(skillvals=skillvals)
    return model.review(vacancy, cv)

@app.get('/health')
async def health_check():
    return {"status": "healthy", "message": "CV analysis service is running"}