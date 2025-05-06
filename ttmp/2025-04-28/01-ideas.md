- Rendering the prompt should be done in the agent, not the command, or at least give the agent a way to add their own variables to render the prompt
- What about that system prompt, do we even need it in the yaml?

- test the file collection with 256 output tokens to cause it to split

- proper document on when to stream output and how to setup the router

- remove the info logging
- add flag to store files on disk
  - agents should be able to add their own parameter layers to the command

  ---

  - [ ] show ai analysis in table. 
  - [ ] hide saved filters and all frmo print view
  - [ ] allow saving custom filters in localstorage
  - [ ] clicking on category / other thing in analysis goes to transactions view
  - [x] clicking on a category also allows filtering.
  - [ ] add "analyze with AI" button
  - [x] sorting of columns
  - [ ] clickable badges for filters
  - [ ] create a "period explanation" as part of the sumary that explains what was going on, same for year