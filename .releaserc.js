// From: https://github.com/asjdf/semantic-release-go-demo
module.exports = {
    branches: ["master", {name: 'alpha', prerelease: true}, {name: 'beta', prerelease: true}],
    plugins: [
        "@semantic-release/commit-analyzer",
        "@semantic-release/release-notes-generator",
        "@semantic-release/github"
    ]
};