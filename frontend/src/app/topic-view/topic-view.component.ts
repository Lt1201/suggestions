import { HttpClient } from '@angular/common/http';
import { Component, Input, QueryList, ViewChildren } from '@angular/core';
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatCardModule } from '@angular/material/card';
import {MatSelectModule} from '@angular/material/select';
import {MatInputModule} from '@angular/material/input';
import {MatFormFieldModule} from '@angular/material/form-field';
import { MatButtonModule } from '@angular/material/button';
import { FormsModule } from '@angular/forms';
import { MatTable, MatTableModule } from '@angular/material/table';
import {MatProgressSpinnerModule} from '@angular/material/progress-spinner';
import { RouterModule } from '@angular/router';

interface suggestion {
  id: number;
  content: string;
  name: string;
  categoryId: number;
}

interface category {
  id: number;
  name: string;
  topicId: number;
  suggestions: suggestion[];
}

interface Topic {
  name: string;
  description: string;
  id: number;
}

interface TopicData {
  topic: Topic;
  categories: category[];
}

@Component({
  selector: 'app-topic-view',
  imports: [MatTooltipModule, MatCardModule, MatInputModule, MatSelectModule, MatFormFieldModule, MatButtonModule, FormsModule, MatTableModule, MatProgressSpinnerModule, RouterModule],
  templateUrl: './topic-view.component.html',
  styleUrl: './topic-view.component.scss'
})
export class TopicViewComponent {
  @Input()
  set id(id: number) {
    this.http.get<TopicData>('/api/topic/'+id).subscribe((topicData) => {
      console.log(topicData);
      this.topic = topicData.topic;
      this.categories = topicData.categories;
      if (this.categories == null) {
        this.categories = [];
      }
      this.isLoading = false;
    });
  }

  @ViewChildren("tables")
  private tables!: QueryList<MatTable<suggestion>>;

  topic: Topic;
  categories: category[] = [];
  isLoading = true;

  newCategoryName = "";
  
  newSuggestionCategoryId = 0;
  newSuggestionName = "";
  newSuggestionContent = "";

  displayedColumns= ['content', 'name'];

  constructor(private http: HttpClient) {
    this.topic = {
      name: "",
      description: "",
      id: 0
    };
  }

  createCategory() {
    console.log("Creating category", this.newCategoryName);
    this.http.post<category>('/api/category', {
      name: this.newCategoryName,
      topicId: this.topic.id
    }).subscribe((category) => {
      category.suggestions = [];
      this.categories.push(category);
      this.newCategoryName = "";
    });
  }

  createSuggestion() {
    console.log("Creating suggestion", this.newSuggestionName, this.newSuggestionContent, this.newSuggestionCategoryId);
    this.http.post<suggestion>('/api/suggestion', {
      name: this.newSuggestionName,
      content: this.newSuggestionContent,
      categoryId: this.newSuggestionCategoryId
    }).subscribe((suggestion) => {
      for (const category of this.categories) {
        if (category.id == this.newSuggestionCategoryId) {
          category.suggestions.push(suggestion);
          this.tables.forEach((table: MatTable<suggestion>) => {
            table.renderRows();
          });
          break;
        }
      }
      this.newSuggestionName = "";
      this.newSuggestionContent = "";
      this.newSuggestionCategoryId = 0;
    });
  }

  suggestionDisabled(){
    return this.newSuggestionName == "" || this.newSuggestionContent == "" || this.newSuggestionCategoryId == 0;
  }

  categoryDisabled(){
    return this.newCategoryName == "";
  }
}
