from openai import OpenAI
from fastapi import FastAPI
from dotenv import load_dotenv
import os

load_dotenv()

app = FastAPI(title='Interview analysis')

PROMPT = open('analys_prompt.txt', 'r').read()
API_TOKEN = os.getenv('PROXY_API_TOKEN')
URL = 'https://api.proxyapi.ru/openai/v1'


client = OpenAI(
    api_key=API_TOKEN,
    base_url=URL,
)


@app.get("/get_review")
async def review_cv(vacancy: str, cv: str, skillvals: str, interview_transcript: str):
    prompt = PROMPT.replace('<vacancy>', vacancy).replace('<resume>', cv).replace('<skill_weights>', skillvals).replace('<interview_transcript>', interview_transcript)
    chat_completion = client.chat.completions.create(
        model="gpt-4o", 
        messages=[
            {
                "role": "user",
                "content": prompt
            }
        ]
    )
    ans = chat_completion.choices[0].message.content
    return {'analysis:': ans}

@app.get('/health')
async def health_check():
    return {"status": "healthy", "message": "Interview analysis service is running"}