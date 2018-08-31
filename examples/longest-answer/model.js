// https://www.javascriptjanuary.com/blog/building-your-first-node-app-using-docker
// docker build -t node-docker .
// docker run --rm node-docker
// or for test
// docker run --rm -e TESTING=true node-docker

const csvdata = require('csvdata');

const testing = 'true' === process.env.TESTING;

// returns the longest answer
const file = testing ? 'swag-dev-head-200.csv' : '/swag.csv';
const resultFile = '/results/predictions.csv';

/* answers is of the form: Array<{ pred: number }>
  This is the expected form to export to leaderboard, e.g.
  [ { pred: 0 },
    { pred: 2 },
    { pred: 3 },
    { pred: 1 },
    ...
  ]
*/
let answers = [];
csvdata.load(file).then(
    (result) => {
        answers = result.map((line => {
            let longestAnswerLength = 0;
            let longestAnswerIndex = 0;
            // line is of the form: { startphrase: string, ending0: string, ending1: string, ending2: string, ending3: string, label: string }
            [line.ending0, line.ending1, line.ending2, line.ending3].forEach((answer, index) => {
                const len = answer.length;
                if (len > longestAnswerLength) {
                    longestAnswerLength = len;
                    longestAnswerIndex = index;
                }
            });
            return {
                pred: longestAnswerIndex
            };
        }));
        csvdata.write(resultFile, answers, {
            header: 'pred'
        })
        if (testing) {
            console.log(answers);
        }
    },
    (error) => console.error(error)
);
