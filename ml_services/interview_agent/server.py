from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from typing import List, Union
from langchain_core.messages import SystemMessage, AIMessage, HumanMessage
from langchain_openai import ChatOpenAI
from langgraph.graph import StateGraph, END, START
from typing import TypedDict
import json
import os
from dotenv import load_dotenv

load_dotenv()

API_TOKEN = os.getenv('PROXY_API_TOKEN')
URL = 'https://api.proxyapi.ru/openai/v1'
CONVERSATION_PATH = 'conversation_history.json'

app = FastAPI(title="Interview Assistant")

class MessageModel(BaseModel):
    type: str  # 'human', 'ai', 'system'
    content: str

class InterviewRequest(BaseModel):
    conversation: List[MessageModel]
    user_input: str

class InterviewResponse(BaseModel):
    conversation: List[MessageModel]
    reply: str


try:
    CV = open('cv_example.txt', 'r', encoding='utf-8').read()
    VACANCY = open('vacancy_example.txt', 'r', encoding='utf-8').read()
    PROMPT = open('interview_prompt.txt', 'r', encoding='utf-8').read()
except FileNotFoundError as e:
    raise HTTPException(status_code=500, detail=f"Required file not found: {e}")

SKILLWEIGHTS = """
    - Релевантный опыт работы: 35%   
    - Ключевые технические навыки: 30%     
    - Географическое соответствие (для очной работы): 15%    
    - Образование: 10%    
    - Soft skills и дополнительные факторы: 10% 
"""

start_prompt = PROMPT.replace('<vacancy>', VACANCY).replace('<resume>', CV).replace('<skills_weights>', SKILLWEIGHTS)

class AgentState(TypedDict):
    messages: List[Union[HumanMessage, AIMessage, SystemMessage]]

model = ChatOpenAI(model='gpt-4o', base_url=URL, openai_api_key=API_TOKEN)

def process(state: AgentState) -> AgentState:
    '''Model check mistackes in CV'''
    prompt = """
    Ты - опытный HR-специалист, который проверяет достоверность информации кандидата.

    РЕЗЮМЕ КАНДИДАТА:
    {CV}

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

graph = StateGraph(AgentState)
graph.add_node('process', process)
graph.add_edge(START, 'process')
graph.add_edge('process', END)
agent = graph.compile()


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


def save_conversation(messages: List[Union[HumanMessage, AIMessage, SystemMessage]]):
    try:
        message_models = to_message_model_list(messages)
        with open(CONVERSATION_PATH, 'w', encoding='utf-8') as f:
            json.dump([msg.model_dump() for msg in message_models], f, ensure_ascii=False, indent=2)
    except Exception as e:
        print(f"Error saving conversation: {e}")

def load_conversation() -> List[MessageModel]:
    try:
        with open(CONVERSATION_PATH, 'r', encoding='utf-8') as f:
            msg_dicts = json.load(f)
            return [MessageModel(**msg) for msg in msg_dicts]
    except FileNotFoundError:
        return [MessageModel(type='system', content=start_prompt)]
    except Exception as e:
        print(f"Error loading conversation: {e}")
        return [MessageModel(type='system', content=start_prompt)]

@app.post('/get_interview_reply', response_model=InterviewResponse)
async def get_interview_reply(request: InterviewRequest):
    """
    Get AI HR response
    """
    try:

        if len(request.conversation) == 1 and request.conversation[0].type == 'string':
            request.conversation[0].type = 'system'
            request.conversation[0].content = start_prompt

        conversation = from_message_model_list(request.conversation)
        
        if not any(isinstance(msg, SystemMessage) for msg in conversation):
            conversation.insert(0, SystemMessage(content=start_prompt))
        
        conversation.append(HumanMessage(content=request.user_input))
        
        result = agent.invoke({'messages': conversation})
        
        save_conversation(result['messages'])
        
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

@app.get('/load_conversation')
async def load_conversation_endpoint():
    conversation = load_conversation()
    return {"conversation": conversation}

@app.delete('/reset_conversation')
async def reset_conversation():
    try:
        if os.path.exists(CONVERSATION_PATH):
            os.remove(CONVERSATION_PATH)
        return {"message": "Conversation history reset successfully"}
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Error resetting conversation: {str(e)}")

@app.get('/health')
async def health_check():
    return {"status": "healthy", "message": "Interview service is running"}