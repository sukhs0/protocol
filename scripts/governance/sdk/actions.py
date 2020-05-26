import json
import sys
from rpc_call import *

class Proposal:
    def __init__(self, pid, pType, description, proposer, init_fund):
        self.pid = pid
        self.pty = pType
        self.des = description
        self.proposer = proposer
        self.init_fund = init_fund

    def _create_proposal(self):
        req = {
            "proposal_id": self.pid,
            "description": self.des,
            "proposer": self.proposer,
            "proposal_type": self.pty,
            "initial_funding": {
                "currency": "OLT",
                "value": convertBigInt(self.init_fund),
            },
            "gasPrice": {
                "currency": "OLT",
                "value": "1000000000",
            },
            "gas": 40000,
        }
        resp = rpc_call('tx.CreateProposal', req)
        print resp
        return resp["result"]["rawTx"]

    def send_create(self):
        # createTx
        raw_txn = self._create_proposal()

        # sign Tx
        signed = sign(raw_txn, self.proposer)

        # broadcast Tx
        result = broadcast_commit(raw_txn, signed['signature']['Signed'], signed['signature']['Signer'])

        if "ok" in result:
            if not result["ok"]:
                sys.exit(-1)
            else:
                print "################### proposal created:" + self.pid
                self.txHash = "0x" + result["txHash"]

    def tx_created(self):
        resp = tx_by_hash(self.txHash)
        return resp["result"]["tx_result"]

class ProposalFund:
    def __init__(self, pid, value, address):
        self.pid = pid
        self.value = value
        self.funder = address

    def _fund_proposal(self):
        req = {
            "proposal_id": self.pid,
            "fund_value": self.value,
            "funder_address": self.funder,
            "gasPrice": {
                "currency": "OLT",
                "value": "1000000000",
            },
            "gas": 40000,
        }
        resp = rpc_call('tx.FundProposal', req)
        print resp
        return resp["result"]["rawTx"]

    def send_fund(self):
        # create Tx
        raw_txn = self._fund_proposal()

        # sign Tx
        signed = sign(raw_txn, self.funder)

        # broadcast Tx
        result = broadcast_commit(raw_txn, signed['signature']['Signed'], signed['signature']['Signer'])

        if "ok" in result:
            if not result["ok"]:
                sys.exit(-1)
            else:
                print "################### proposal funded:" + Proposal
                return result["txHash"]

class ProposalVote:
    def __init__(self, pid, opinion, address):
        self.pid = pid
        self.opin = opinion
        self.voter = address

    def _vote_proposal(self):
        req = {
            "proposal_id": self.pid,
            "opinion": self.opin,
            "gasPrice": {
                "currency": "OLT",
                "value": "1000000000",
            },
            "gas": 40000,
        }
        resp = rpc_call('tx.VoteProposal', req, self.voter)
        result = resp["result"]
        print resp
        return result["rawTx"], result['signature']['Signed'], result['signature']['Signer']

    def send_vote(self):
        # create and sign Tx
        raw_txn, signed, signer = self._vote_proposal()

        # broadcast Tx
        result = broadcast_commit(signed["rawTx"], signed['signature']['Signed'], signed['signature']['Signer'])
        
        if "ok" in result:
            if not result["ok"]:
                sys.exit(-1)
            else:
                print "################### proposal voted:" + self.pid + "opinion: " + self.opin
                return result["txHash"]

def addresses():
    resp = rpc_call('owner.ListAccountAddresses', {})
    return resp["result"]["addresses"]


def sign(raw_tx, address):
    resp = rpc_call('owner.SignWithAddress', {"rawTx": raw_tx, "address": address})
    return resp["result"]


def broadcast_commit(raw_tx, signature, pub_key):
    resp = rpc_call('broadcast.TxCommit', {
        "rawTx": raw_tx,
        "signature": signature,
        "publicKey": pub_key,
    })
    print resp
    if "result" in resp:
        return resp["result"]
    else:
        return resp


def broadcast_sync(raw_tx, signature, pub_key):
    resp = rpc_call('broadcast.TxSync', {
        "rawTx": raw_tx,
        "signature": signature,
        "publicKey": pub_key,
    })
    return resp["result"]

def query_proposals(prefix):
    req = {
        "prefix": prefix,
        "gasPrice":
        {
            "currency": "OLT",
            "value": "1000000000",
        },
        "gas": 40000,
    }

    resp = rpc_call('query.GetProposals', req)
    print json.dumps(resp, indent=4)
    return resp["result"]["proposals"]

def query_proposal(proposal_id):
    req = {"proposal_id": proposal_id}
    resp = rpc_call('query.GetProposalByID', req)
    print json.dumps(resp, indent=4)
    return resp["result"]
