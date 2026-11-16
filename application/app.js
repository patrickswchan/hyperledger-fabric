/*
 * Client for the invoice chaincode.
 *
 * A Fabric Gateway client keeps one gRPC connection to a peer. That peer's
 * gateway service collects endorsements from the other organizations, submits
 * the transaction to the orderer, and reports when it commits.
 *
 * Submit transactions write to the ledger (Issue, Accept, Pay, Cancel).
 * Evaluate transactions only read (Read, List, History).
 */

const grpc = require('@grpc/grpc-js');
const { connect, hash, signers } = require('@hyperledger/fabric-gateway');
const crypto = require('node:crypto');
const fs = require('node:fs/promises');
const path = require('node:path');

const channelName = envOrDefault('CHANNEL_NAME', 'mychannel');
const chaincodeName = envOrDefault('CHAINCODE_NAME', 'invoice');
const peerEndpoint = envOrDefault('PEER_ENDPOINT', 'localhost:7051');
const peerHostAlias = envOrDefault('PEER_HOST_ALIAS', 'peer0.org1.example.com');

const repoRoot = path.resolve(__dirname, '..');
const cryptoRoot = envOrDefault(
    'TEST_NETWORK_ORG_PATH',
    path.join(
        repoRoot,
        '.fabric',
        'fabric-samples',
        'test-network',
        'organizations',
        'peerOrganizations',
    ),
);

const org1 = organization('Org1MSP', 'org1.example.com', 'User1@org1.example.com');
const org2 = organization('Org2MSP', 'org2.example.com', 'User1@org2.example.com');

// Both identities dial Org1's peer. TLS checks that peer's certificate, so
// the CA is Org1's even when the signer is an Org2 user.
const gatewayTlsCertPath = path.join(
    cryptoRoot,
    'org1.example.com',
    'peers',
    'peer0.org1.example.com',
    'tls',
    'ca.crt',
);

function organization(mspId, domain, user) {
    const orgPath = path.join(cryptoRoot, domain);
    return {
        mspId,
        certDirectory: path.join(orgPath, 'users', user, 'msp', 'signcerts'),
        keyDirectory: path.join(orgPath, 'users', user, 'msp', 'keystore'),
    };
}

async function main() {
    const invoiceId = envOrDefault('INVOICE_ID', `inv-${Date.now()}`);
    console.log(`Channel ${channelName}, chaincode ${chaincodeName}, invoice ${invoiceId}`);

    // Both identities dial the same gateway peer. The signer certificate,
    // not the peer they dial, decides which organization the transaction
    // is submitted as.
    const org1Gateway = await connectAs(org1);
    const org2Gateway = await connectAs(org2);

    try {
        const issuer = org1Gateway.gateway.getNetwork(channelName).getContract(chaincodeName);
        const payer = org2Gateway.gateway.getNetwork(channelName).getContract(chaincodeName);

        console.log('\n--> Submit Issue as Org1MSP (the issuer)');
        const issued = await submit(issuer, 'Issue', invoiceId, 'Org2MSP', 'USD', 'Steel coil shipment', '125000');
        console.log(`*** Committed\n${issued}`);

        console.log('\n--> Evaluate Read (no ledger write)');
        console.log(await evaluate(issuer, 'Read', invoiceId));

        console.log('\n--> Submit Accept as Org1MSP');
        await expectChaincodeError(() => submit(issuer, 'Accept', invoiceId));

        console.log('\n--> Submit Accept as Org2MSP (the payer)');
        console.log(`*** Committed\n${await submit(payer, 'Accept', invoiceId)}`);

        console.log('\n--> Submit Pay as Org1MSP');
        await expectChaincodeError(() => submit(issuer, 'Pay', invoiceId));

        console.log('\n--> Submit Pay as Org2MSP');
        console.log(`*** Committed\n${await submit(payer, 'Pay', invoiceId)}`);

        console.log('\n--> Submit Cancel of a paid invoice as Org1MSP');
        await expectChaincodeError(() => submit(issuer, 'Cancel', invoiceId));

        const cancelledId = `${invoiceId}-void`;
        console.log(`\n--> Issue ${cancelledId} and cancel it as the issuer`);
        await submit(issuer, 'Issue', cancelledId, 'Org2MSP', 'USD', 'Order withdrawn', '500');
        console.log(`*** Committed\n${await submit(issuer, 'Cancel', cancelledId)}`);

        console.log('\n--> Evaluate History');
        console.log(await evaluate(issuer, 'History', invoiceId));

        console.log('\n--> Evaluate List');
        console.log(await evaluate(payer, 'List'));
    } finally {
        org1Gateway.close();
        org2Gateway.close();
    }
}

async function connectAs(org) {
    const client = await newGrpcConnection(gatewayTlsCertPath);
    const gateway = connect({
        client,
        identity: await newIdentity(org),
        signer: await newSigner(org),
        hash: hash.sha256,
        evaluateOptions: () => ({ deadline: Date.now() + 5000 }),
        endorseOptions: () => ({ deadline: Date.now() + 15000 }),
        submitOptions: () => ({ deadline: Date.now() + 5000 }),
        commitStatusOptions: () => ({ deadline: Date.now() + 60000 }),
    });
    return {
        gateway,
        close() {
            gateway.close();
            client.close();
        },
    };
}

async function newGrpcConnection(tlsCertPath) {
    const tlsRootCert = await fs.readFile(tlsCertPath);
    const tlsCredentials = grpc.credentials.createSsl(tlsRootCert);
    return new grpc.Client(peerEndpoint, tlsCredentials, {
        'grpc.ssl_target_name_override': peerHostAlias,
    });
}

async function newIdentity(org) {
    const certPath = await firstFile(org.certDirectory);
    return {
        mspId: org.mspId,
        credentials: await fs.readFile(certPath),
    };
}

async function newSigner(org) {
    const keyPath = await firstFile(org.keyDirectory);
    const privateKeyPem = await fs.readFile(keyPath);
    const privateKey = crypto.createPrivateKey(privateKeyPem);
    return signers.newPrivateKeySigner(privateKey);
}

async function firstFile(directory) {
    let names;
    try {
        names = await fs.readdir(directory);
    } catch (error) {
        throw new Error(
            `Cannot read ${directory}. Start the test network with Fabric CAs first: ./scripts/up.sh\n${error.message}`,
        );
    }
    const file = names[0];
    if (!file) {
        throw new Error(`No enrollment material in ${directory}`);
    }
    return path.join(directory, file);
}

async function submit(contract, fn, ...args) {
    const result = await contract.submitTransaction(fn, ...args);
    return decode(result);
}

async function evaluate(contract, fn, ...args) {
    const result = await contract.evaluateTransaction(fn, ...args);
    return decode(result);
}

function decode(bytes) {
    const text = Buffer.from(bytes).toString('utf8');
    if (!text) {
        return '';
    }
    try {
        return JSON.stringify(JSON.parse(text), null, 2);
    } catch {
        return text;
    }
}

async function expectChaincodeError(fn) {
    try {
        await fn();
    } catch (error) {
        console.log(`*** Rejected, as intended: ${errorText(error)}`);
        return;
    }
    throw new Error('transaction succeeded but the chaincode should have rejected it');
}

function errorText(error) {
    const parts = [error.message];
    const details = error.details;
    if (Array.isArray(details)) {
        for (const detail of details) {
            if (detail && typeof detail.message === 'string') {
                parts.push(detail.message);
            }
        }
    }
    return parts.join('\n');
}

function envOrDefault(key, fallback) {
    const value = process.env[key];
    return value && value.length > 0 ? value : fallback;
}

main().catch((error) => {
    console.error('\nApplication failed:', error);
    process.exitCode = 1;
});
