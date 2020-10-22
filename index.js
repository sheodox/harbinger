console.log('Remote Log Collector started!')

const FETCH_INTERVAL_MS = 10000,
	{servers} = require('./config.json'),
	fetch = require('node-fetch'),
	{Client} = require('@elastic/elasticsearch'),
	elasticClient = new Client({
		node: 'http://es01:9200'
	});

const createPromises = [];
for (const {name} of servers) {
	createPromises.push(new Promise((resolve, reject) => {
		elasticClient.indices.create({
			index: name
		}, (err) => {
			!err || err.message === 'resource_already_exists_exception' ? resolve() : reject(err);
		});
	}));
}
const indicesReady = Promise.all(createPromises);

async function fetchLogs() {
	await indicesReady;

	for (const {path, bearer, name} of servers) {
		try {
			//collect a batch of logs
			const logs = await fetch(path, {
				headers: {
					Authorization: `Bearer ${bearer}`,
					'User-Agent': 'remote-log-collector'
				}
			}).then(res => res.json());


			// if there were any logs, push them into elasticsearch
			if (logs.length) {
				console.log(`Received ${logs.length} logs from ${path}`);
				const body = logs.flatMap(doc => [{index: {_index: name}}, doc])
				await elasticClient.bulk({
					body
				})
			}
		}
		catch(e) {
			console.error(`Error fetching logs from "${name}"`, e);
		}
	}
}

setInterval(fetchLogs, FETCH_INTERVAL_MS);