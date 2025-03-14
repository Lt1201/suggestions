import { HttpClient } from '@angular/common/http';
import { Component } from '@angular/core';
import { MatListModule } from '@angular/material/list';
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatCardModule } from '@angular/material/card';
import {MatSelectModule} from '@angular/material/select';
import {MatInputModule} from '@angular/material/input';
import {MatFormFieldModule} from '@angular/material/form-field';
import { MatButtonModule } from '@angular/material/button';
import { FormsModule } from '@angular/forms';
import { RouterLink, RouterLinkActive } from '@angular/router';

interface Topic {
  name: string;
  description: string;
  id: number;
}

@Component({
  selector: 'app-topics',
  imports: [MatListModule, MatTooltipModule, MatCardModule, MatInputModule, MatSelectModule, MatFormFieldModule, MatButtonModule, FormsModule, RouterLink, RouterLinkActive],
  templateUrl: './topics.component.html',
  styleUrl: './topics.component.scss'
})
export class TopicsComponent {
  topics: Topic[] = [];
  newTopicName = "";
  newTopicDescription = "";

  constructor(private http: HttpClient) {
    http.get<Topic[]>('/api/topic').subscribe((topics) => {
      console.log(topics);
      this.topics = topics;
    });
  }

  createTopic() {
    console.log("Creating topic", this.newTopicName, this.newTopicDescription);
    this.http.post<Topic>('/api/topic', {
      name: this.newTopicName,
      description: this.newTopicDescription
    }).subscribe((topic) => {
      this.topics.push(topic);
    });
  }
}
