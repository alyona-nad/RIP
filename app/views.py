from django.shortcuts import render

# Create your views here.
from rest_framework.decorators import api_view
from rest_framework.response import Response
from rest_framework import status

import time
import random
import requests

from concurrent import futures

CALLBACK_URL = "http://127.0.0.1:8080/update_dyes/"

executor = futures.ThreadPoolExecutor(max_workers=1)

def get_random_status(pk):
    time.sleep(5)
    return {
      "ID_Dye": pk,
      "Price": int(random.randint(500,1000)),
      "Key":"123456",
    }

def status_callback(task):
    try:
      result = task.result()
      print(result)
    except futures._base.CancelledError:
      return
    
    nurl = str(CALLBACK_URL+str(result["ID_Dye"])+'/put/')
    answer = {"Price": result["Price"],"Key":"123456"}
    print(answer)
    requests.put(nurl, json=answer, timeout=3)

@api_view(['POST'])
def set_status(request):
    if "pk" in request.data.keys():   
        id = request.data["pk"]        

        task = executor.submit(get_random_status, id)
        task.add_done_callback(status_callback)        
        return Response(status=status.HTTP_200_OK)
    return Response(status=status.HTTP_400_BAD_REQUEST)
