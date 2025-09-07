from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from openai import OpenAI
from typing import List, Union
from langchain_core.messages import SystemMessage, AIMessage, HumanMessage
from langchain_openai import ChatOpenAI
from langgraph.graph import StateGraph, END, START
from typing import TypedDict, Literal
import json
import os
from dotenv import load_dotenv

#configs
try:
    load_dotenv()
    API_TOKEN = os.getenv('PROXY_API_TOKEN')
    CV = open('cv_example.txt', 'r', encoding='utf-8').read()
    VACANCY = open('vacancy_example.txt', 'r', encoding='utf-8').read()
    INTERVIEW_PROMPT = open('interview_prompt.txt', 'r', encoding='utf-8').read()
    ANALYSIS_PROMPT = open('analys_prompt.txt', 'r').read()
except FileNotFoundError as e:
    raise HTTPException(status_code=500, detail=f"Required file not found: {e}")

URL = 'https://api.proxyapi.ru/openai/v1'

SKILLWEIGHTS = """
    - Релевантный опыт работы: 35%   
    - Ключевые технические навыки: 30%     
    - Географическое соответствие (для очной работы): 15%    
    - Образование: 10%    
    - Soft skills и дополнительные факторы: 10% 
"""

# set up
client = OpenAI(
    api_key=API_TOKEN,
    base_url=URL,
)

app = FastAPI(title="Interview Assistant")


# fields
class MessageModel(BaseModel):
    type: str  # 'human', 'ai', 'system'
    content: str

class InterviewRequest(BaseModel):
    conversation: List[MessageModel]
    user_input: str

class InterviewResponse(BaseModel):
    conversation: List[MessageModel]
    reply: str


class AgentState(TypedDict):
    messages: List[Union[HumanMessage, AIMessage, SystemMessage]]


class IsEnd(BaseModel):
    choice: Literal['end', 'continue']

#convert fields
def from_message_model_list(msg_models: List[MessageModel]) -> List[Union[HumanMessage, AIMessage, SystemMessage]]:
    msg_list = []
    for msg in msg_models:
        if msg.type == 'human':
            msg_list.append(HumanMessage(content=msg.content))
        elif msg.type == 'ai':
            msg_list.append(AIMessage(content=msg.content))
        elif msg.type == 'system':
            msg_list.append(SystemMessage(content=msg.content))
        else:
            raise HTTPException(status_code=400, detail=f"Unsupported message type: {msg.type}")
    return msg_list

def to_message_model_list(msg_list: List[Union[HumanMessage, AIMessage, SystemMessage]]) -> List[MessageModel]:
    models = []
    for msg in msg_list:
        if isinstance(msg, HumanMessage):
            models.append(MessageModel(type='human', content=msg.content))
        elif isinstance(msg, AIMessage):
            models.append(MessageModel(type='ai', content=msg.content))
        elif isinstance(msg, SystemMessage):
            models.append(MessageModel(type='system', content=msg.content))
    return models

def messages_to_text(msg_list: List[MessageModel]) -> str:
    text = ""
    for msg in msg_list:
        if msg.type == 'human':
            text += f"Кандидат: {msg.content}\n"
        elif msg.type == 'ai':
            text += f"HR: {msg.content}\n"
        elif msg.type == 'system':
            text += f"Система: {msg.content}\n"
    return text.strip()



#agent

model = ChatOpenAI(model='gpt-4o', base_url=URL, openai_api_key=API_TOKEN)


def decide_next_node(state:AgentState) -> AgentState:
    '''This node will select the next node of the graph'''

    prompt = """ 
    Определи это завершение диалога или нет.     
    Ответ должен быть один из двух: 
    end - если завершение диалога
    continue - если продолжение диалога"""

    structured_response_model = model.with_structured_output(IsEnd)
    response = structured_response_model.invoke(state['messages'] + [prompt])

    return response.choice


def process(state: AgentState) -> AgentState:
    '''Model check mistackes in CV'''
    prompt = """
    Ты - опытный HR-специалист, который проверяет достоверность информации кандидата.

    РЕЗЮМЕ КАНДИДАТА В НАЧАЛЕ В СИСТЕМНОМ ПРОМПТЕ

    ТВОЯ ЗАДАЧА:
    Сравни ответ кандидата с информацией в его резюме и определи наличие несоответствий.

    КРИТЕРИИ АНАЛИЗА:
    1. Даты и временные рамки: Есть ли противоречия в хронологии проектов и стажа относительно его ответа и резюме?
    2. Опыт работы: Противоречат ли сроки работы, должности или обязанности тем, что указаны в CV?
    3. Технические навыки: Кандидат не упоминает технологии которые заявлены в резюме и подразумевались в его ответе?
    4. Образование: Соответствует ли уровень знаний заявленному образованию?

    ОСОБЕННОСТИ ОЦЕНКИ:
    - Путаница в датах или компаниях - СЕРЬЕЗНОЕ несоответствие
    - Незначительные расширения информации (детали, которые логично не включать в резюме) - НЕ считаются несоответствием
    - Преувеличение роли или ответственности - СЧИТАЕТСЯ несоответствием
    - Упоминание новых технологий без объяснения, где их изучил - ПОДОЗРИТЕЛЬНО, СТОИТ СПРОСИТЬ ОБ ЭТОМ

    ПРИМЕР АНАЛИЗА:
    Если кандидат говорит о 5-летнем опыте, а в резюме указано 3 года - серьезная проблема.
    Если кандидат говорит, что хорошо владеет React, а в резюме только Vue - это подозрительно.
    Если кандидат детализирует проект из резюме - это нормально.

    ЗАДАВАЙ УТОЧНЯЮЩИЕ ВОПРОСЫ ЕСЛИ ЕСТЬ СОМНЕНИЯ
    ИМЕЙ ВВИДУ, ЧТО ТЫ ОТВЕЧАЕШЬ КАНДИДАТУ И ЕМУ НУЖНО ОТВЕЧАТЬ КАК ЧЕЛОВЕК ЧЕЛОВЕКУ, А НЕ ВЫДАВАТЬ СТАТИСТИКУ ПО ЕГО ОШИБКАМ
    НЕ НУЖНО ПИСАТЬ ЕМУ ЧТО ЕГО ДАННЫЕ СООТВЕТСТВУЮТ РЕЗЮМЕ И ПОДОБНЫЕ ФРАЗЫ, ОНИ ДАЮТ ПОНЯТЬ ЧТО ТЫ СВЕРЯЕШЬ ЕГО ОТВЕТЫ С РЕЗЮМЕ, ТО ЕСТЬ ПИСАТЬ ПОДОБНЫЕ ВЕЩИ НЕ НАДО: "Вы упомянули, что провели 3 месяца на позиции ML стажёра. Это соответствует вашему резюме, где указано, что вы были ML стажёром в компании Webee в 2025 году."
    НЕ НУЖНО "ВЪЕДАТЬСЯ" В КАНДИДАТА БЕСКОНЕЧНЫМИ ВОПРОСАМИ О ЕГО НЕСОСТЫКОВКАХ ЕСЛИ ОН НЕ СМОГ НОРМАЛЬНО ОБЪЯСНИТЬСЯ ИДИ ДАЛЬШЕ НЕ ПЕРЕЗАДАВАЙ ТОТ ЖЕ ВОПРОС

    ЕСЛИ ВСЕ В ПОРЯДКЕ ПРОДОЛЖАЙ ИНТЕРВЬЮ

    ВСЕ ТВОИ ОТВЕТЫ ОЗВУЧИВАЮТСЯ ГОЛОСОМ И ОТПРАВЛЯЮТСЯ ПОЛЬЗОВАТЕЛЮ, ПОЭТОМУ СТАРАЙСЯ ТАК ЧТОБЫ БЫЛО УДОБНО ОЗВУЧИВАТЬ, БУДЬ ПРОФЕССИОНАЛЬНЫМ HR!!
    """
    response = model.invoke(state['messages'] + [prompt])
    state['messages'].append(AIMessage(content=response.content))
    return state

def get_review(state: AgentState) -> AgentState:
    '''End the dialoge and get review'''
    cv_vac_skills = state['messages'][0].content.split('<sep>')[1]
    interview_transcript = to_message_model_list(state['messages'])
    try:
        prompt = ANALYSIS_PROMPT.replace('<interview_transcript>', messages_to_text(interview_transcript)).replace('cv_vac_skills', cv_vac_skills)
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
        state['messages'].append(AIMessage(content=ans))
        return state
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Internal server error: {str(e)}")

graph = StateGraph(AgentState)

graph.add_node('router', lambda state:state)
graph.add_node('process', process)
graph.add_node('get_review', get_review)
graph.add_conditional_edges(
    'router',
    decide_next_node,
    {
        'continue': 'process',
        'end': 'get_review'
    }
)


graph.add_edge(START, 'router')
graph.add_edge('process', END)
graph.add_edge('get_review', END)
agent = graph.compile()


#endpoints
@app.post('/get_interview_reply', response_model=InterviewResponse)
async def get_interview_reply(request: InterviewRequest):
    '''Get AI HR response in dialoge'''
    try:

        # start a dialoge
        # VACANCY, CV, SKILL_WEIGHTS must be from broker
        if len(request.conversation) == 1 and request.conversation[0].type == 'string':
            start_prompt = INTERVIEW_PROMPT.replace('<vacancy>', VACANCY).replace('<resume>', CV).replace('<skills_weights>', SKILLWEIGHTS)
            request.conversation[0].type = 'system'
            request.conversation[0].content = start_prompt

        conversation = from_message_model_list(request.conversation)
        
        conversation.append(HumanMessage(content=request.user_input))
        
        result = agent.invoke({'messages': conversation})
                
        last_ai_message = None
        for msg in reversed(result['messages']):
            if isinstance(msg, AIMessage):
                last_ai_message = msg
                break
        
        reply = last_ai_message.content if last_ai_message else "Извините, произошла ошибка при генерации ответа."
        
        response_models = to_message_model_list(result['messages'])
        
        return InterviewResponse(
            conversation=response_models,
            reply=reply
        )
        
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Internal server error: {str(e)}")

@app.get('/health')
async def health_check():
    return {"status": "healthy", "message": "Interview service is running"}