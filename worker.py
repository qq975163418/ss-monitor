#!/usr/bin/python
# -*- coding: utf8 -*-
from urllib.request import Request, urlopen
from configparser import ConfigParser
import json
import os
import logging
import traceback
import sys
from urllib.parse import urlencode
from httpstatus import HTTPStatus


class DictWrapper(dict):
    """ dictインスタンスへインスタンス変数を追加するために使用する """
    pass


class BytesWrapper(bytes):
    """ bytesインスタンスへインスタンス変数を追加するために使用する """
    pass


class Worker:
    def __init__(self, filepath):
        path = os.path.split(os.path.realpath(filepath))[0] + '/config.ini'
        self.config = ConfigParser()
        try:
            self.config.read(path)
        except:
            logging.critical('Fail to read config file %s' % path)
            traceback.print_exc(file=sys.stderr)
            exit(-1)
        self.type = ""
        self.apiUrl = self.config.get('worker', 'apiUrl')
        self.apiKey = self.config.get('worker', 'apiKey')
        self.name = self.config.get('worker', 'name')
        self.maxThreads = int(self.config.get('worker', 'maxThreads'))
        self.heartRate = int(self.config.get('worker', 'heartRate'))
        self.threads = []
        self.threadsCnt = 0
        self.maxTry = int(self.config.get('worker', 'maxTry'))
        self.connTimeout = int(self.config.get('worker', 'connTimeout'))
        self.loggingLevel = self.config.get('worker', 'loggingLevel')
        self.loggingLevelDict = {
            'NOTSET': logging.NOTSET,
            'DEBUG': logging.DEBUG,
            'INFO': logging.INFO,
            'WARNING': logging.WARNING,
            'ERROR': logging.ERROR,
            'CRITICAL': logging.CRITICAL
        }
        logging.basicConfig(level=self.loggingLevelDict[self.loggingLevel],
                            stream=sys.stdout,
                            format='[%(asctime)s] %(threadName)s %(filename)s [%(levelname)s] %(message)s',
                            datefmt='%Y %b %d %H:%M:%S', )

    def _GET(self, path, data_dict=None, isDeserialize=True, headers=None, method='GET', depth=0):
        data = urlencode(data_dict)
        if headers is None:
            headers = {'X-Access-Key': self.apiKey}
        else:
            headers['X-Access-Key'] = self.apiKey
        req = Request(
            url=self.apiUrl + ('/' + path if path else '') + '?' + data,
            method=method,
            headers=headers
        )
        try:
            with urlopen(req, timeout=self.connTimeout) as res:
                resbin = res.read()
                if isDeserialize:
                    data = DictWrapper(json.loads(str(resbin, 'utf8')))
                else:
                    data = BytesWrapper(resbin)
                # HTTPステータスコードとヘッダーを追加
                data.code = res.code
                data.msg = res.msg
                data.headers = res.headers
                return data
        except:
            if depth < self.maxTry:
                logging.warning("Connection failed after trying %d times, reconnecting..." % (depth + 1))
                traceback.print_exc(file=sys.stderr)
                return self._GET(path, data_dict, isDeserialize, headers, method, depth + 1)
            else:
                logging.error("Connection failed after trying %d times, exiting..." % self.maxTry)
                logging.error("Data: %s" % data)
                traceback.print_exc(file=sys.stderr)
                return

    def _POST(self, path, data_dict=None, isDeserialize=True, headers=None, method='POST', depth=0):
        data = urlencode(data_dict)
        if headers is None:
            headers = {'X-Access-Key': self.apiKey}
        else:
            headers['X-Access-Key'] = self.apiKey
        req = Request(
            url=self.apiUrl + ('/' + path if path else ''),
            method=method,
            headers=headers,
            data=bytes(data, 'utf8')
        )
        try:
            with urlopen(req, timeout=self.connTimeout) as res:
                resbin = res.read()
                if isDeserialize:
                    data = DictWrapper(json.loads(str(resbin, 'utf8')))
                else:
                    data = BytesWrapper(resbin)
                # HTTPステータスコードとヘッダーを追加
                data.code = res.code
                data.msg = res.msg
                data.headers = res.headers
                return data
        except:
            if depth < self.maxTry:
                logging.warning("Connection failed after trying %d times, reconnecting..." % (depth + 1))
                traceback.print_exc(file=sys.stderr)
                return self._POST(path, data_dict, isDeserialize, headers, method, depth + 1)
            else:
                logging.error("Connection failed after trying %d times, exiting..." % self.maxTry)
                logging.error("Data: %s" % data)
                traceback.print_exc(file=sys.stderr)
                return

    def _DELETE(self, *args, **nargs):
        return self._GET(*args, method='DELETE', **nargs)

    def _PUT(self, path, data_dict=None, isDeserialize=True, headers=None):
        return self._POST(path, data_dict, isDeserialize, headers=headers, method='PUT')

    def callback(self, task):
        try:
            logging.info('Calling back task %s' % task['ID'])
            rst = self._POST(path='task/' + str(task['ID']) + '/callback', data_dict={'worker': self.name})
            if not (rst.code == HTTPStatus.OK and rst['result']):
                logging.error('Callback task result: %s' % rst)
                raise Exception('Failed to callback task')
        except:
            logging.error('Failed to callback task %s' % task['ID'])
            traceback.print_exc(file=sys.stderr)
            return False
        return True
