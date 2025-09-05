import requests
from openai import OpenAI
from abc import ABC
from abc import abstractmethod
from dotenv import load_dotenv
import os

load_dotenv()

TOKEN = os.getenv('TOKEN')
PROXY_API_TOKEN = os.getenv('PROXY_API_TOKEN')
# TOKEN = 'io-v2-eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJvd25lciI6IjExZGMxM2RmLWFiMDMtNGE1NC05ZDkzLTY1MDhiZTJjYTgwNyIsImV4cCI6NDkxMDA5MjMyMH0.CzUoII1RDO_nnyiQRyKWZeZwYjbJggQbO-_IvX1ZFRfn1UZHs3VEcytED8CpVg6q6phdWsHMy-FCHIxyXZJ3ZQ'
URL = "https://api.proxyapi.ru/openai/v1"
PROMPT = open('cv_review_prompt.txt', 'r').read()

class ReviewerModel(ABC):
    def __init__(self, token: str, skillvals: str = None, system_prompt: str = PROMPT):
        self.token = token
        self.system_prompt = system_prompt
        self.skillvals = skillvals
        if not skillvals:
            raise ValueError("Skillvals can not be None")

    @abstractmethod
    def review(self, cv: str):
        pass

class GPTReviewer(ReviewerModel):
    def __init__(self, model_name: str = "gpt-4o", token: str = PROXY_API_TOKEN, skillvals: str = None, url: str = URL, system_prompt: str = PROMPT):
        super().__init__(token, skillvals, system_prompt)
        self.url = url
        self.model_name = model_name
        self.system_prompt = self.system_prompt.replace('<skillval>', skillvals)
        self.client = OpenAI(
            api_key=self.token,
            base_url=self.url,
        )
    
    def review(self, vacancy: str, cv: str):        
        self.system_prompt = self.system_prompt.replace('<vac>', vacancy)
        chat_completion = self.client.chat.completions.create(model=self.model_name, 
                                                              messages=[
                                                                {
                                                                    "role": "system",
                                                                    "content": self.system_prompt
                                                                },
                                                                {      
                                                                    "role": "user",
                                                                    "content": f'РЕЗЮМЕ\n{cv}'
                                                                }])
        # headers = {
        #     "Content-Type": "application/json",
        #     "Authorization": f"Bearer {self.token}"
        # }

        # data = {
        #     "model": "openai/gpt-oss-120b",
        #     "messages": [
        #         {
        #             "role": "system",
        #             "content": 'self.system_prompt'
        #         },
        #         {
        #             "role": "user",
        #             "content": f'РЕЗЮМЕ\n{cv}'
        #         }
        #     ]
        # }

        # response = requests.post(self.url, headers=headers, json=data)
        # print(response.json())
        # ans = response.json()['choices'][0]['message']['content']
        ans = chat_completion.choices[0].message.content
        return ans